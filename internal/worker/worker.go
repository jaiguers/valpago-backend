package worker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/usuario/valpago-backend/internal/config"
	"github.com/usuario/valpago-backend/internal/db"
)

func Start() {
	if db.Rdb == nil {
		log.Println("Redis not available, skipping worker...")
		return
	}

	log.Println("Starting Redis worker...")

	ctx := context.Background()
	groupName := config.C.RedisGroup
	streamName := config.C.RedisStreamNS
	consumerName := fmt.Sprintf("%s-%d", config.C.RedisConsumer, time.Now().Unix())

	// Create consumer group if it doesn't exist
	err := db.Rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Printf("Error creating consumer group %s for stream %s: %v", groupName, streamName, err)
	}

	for {
		// Read from Redis stream
		streams, err := db.Rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    time.Second * 5,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				// No messages, continue
				continue
			}
			log.Printf("Redis stream error: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				// Process transaction
				processTransaction(ctx, message)

				// Acknowledge message
				db.Rdb.XAck(ctx, streamName, groupName, message.ID)
			}
		}
	}
}

func processTransaction(ctx context.Context, message redis.XMessage) {
	log.Printf("Processing transaction: %s", message.ID)

	// Simulate processing time
	// time.Sleep(time.Second * 2)

	// Notificar al front solo los PENDING (para que aparezcan en la cola)
	if raw, ok := message.Values["data"]; ok {
		originalJSON := fmt.Sprintf("%v", raw)

		// 1) Intentar descargar imagen de Meta y subir a Supabase
		uploadedURL, err := fetchAndUploadSupportImage(ctx, originalJSON)
		if err != nil {
			log.Printf("support image handling error: %v", err)
		}

		// 2) Si tenemos URL subida, actualizar primero en Mongo
		var finalJSON string = originalJSON
		if uploadedURL != "" {
			idStr, idErr := extractIdFromTransactionJSON(originalJSON)
			if idErr != nil {
				log.Printf("failed to extract id: %v", idErr)
			} else {
				if err := setMongoSupportURL(ctx, idStr, uploadedURL); err != nil {
					log.Printf("failed to update Mongo support_url: %v", err)
				} else {
					// 3) Recargar la transacción desde Mongo y usar ese JSON actualizado para notificar
					var txDoc map[string]interface{}
					oid, _ := primitive.ObjectIDFromHex(idStr)
					if err := db.Mongo().Collection("transactions").FindOne(ctx, bson.M{"_id": oid}).Decode(&txDoc); err == nil {
						if b, mErr := json.Marshal(txDoc); mErr == nil {
							finalJSON = string(b)
						}
					}
				}
			}
		}

		// 4) Publicar notificación PENDING con el JSON final (ya con support_url de Supabase si se logró subir)
		notification := map[string]interface{}{
			"type":      "transaction.pending",
			"data":      finalJSON,
			"timestamp": time.Now().Unix(),
		}
		db.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: config.C.RedisNotificationsStream,
			Values: notification,
		})
		return
	}

	// Manejar otros tipos de eventos (review, approved, rejected)
	if eventType, ok := message.Values["type"]; ok {
		eventTypeStr := fmt.Sprintf("%v", eventType)
		payload := ""
		if raw, hasData := message.Values["data"]; hasData {
			payload = fmt.Sprintf("%v", raw)
		}

		// Notificar al front según el tipo de evento
		notification := map[string]interface{}{
			"type":      eventTypeStr,
			"data":      payload,
			"timestamp": time.Now().Unix(),
		}
		db.Rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: config.C.RedisNotificationsStream,
			Values: notification,
		})
		return
	}

	// Fallback: publicar evento mínimo si no hay "data"
	notificationData := map[string]interface{}{
		"status":    "review",
		"timestamp": time.Now().Unix(),
	}
	db.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: config.C.RedisNotificationsStream,
		Values: notificationData,
	})
}

// fetchAndUploadSupportImage intenta descargar la imagen desde Meta con Bearer y subirla a Supabase; si falla, usa imagen local
func fetchAndUploadSupportImage(ctx context.Context, transactionJSON string) (string, error) {
	// Extraer support_url del JSON (búsqueda simple para evitar dependencia de structs)
	// Se asume que el campo es: "support_url":"<url>"
	idx := strings.Index(transactionJSON, "\"support_url\":\"")
	if idx == -1 {
		return "", fmt.Errorf("support_url not found in payload")
	}
	start := idx + len("\"support_url\":\"")
	end := strings.Index(transactionJSON[start:], "\"")
	if end == -1 {
		return "", fmt.Errorf("malformed support_url in payload")
	}
	supportURL := transactionJSON[start : start+end]

	// Obtener URL real de la imagen desde Graph API
	realImageURL, err := getRealImageURLFromMeta(ctx, supportURL)
	if err != nil {
		log.Printf("Error getting real image URL from Meta: %v", err)
		return fallbackUpload()
	}

	// Descargar imagen real con Bearer
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realImageURL, nil)
	if err != nil {
		log.Println(err, "error NewRequestWithContext descargando imagen")
		return fallbackUpload()
	}
	if config.C.BearerTokenMeta != "" {
		req.Header.Set("Authorization", "Bearer "+config.C.BearerTokenMeta)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp != nil {
			resp.Body.Close()
		}
		log.Println(err, "error DefaultClient.Do imagen")
		log.Println(resp.StatusCode, "status code imagen")
		log.Println(resp.Status, "status imagen")
		return fallbackUpload()
	}
	defer resp.Body.Close()

	// Asegurar que es imagen
	contentType := resp.Header.Get("Content-Type")

	if contentType == "" {
		// intentar detectar por algunos bytes
		peek := make([]byte, 512)
		n, _ := io.ReadFull(resp.Body, peek)
		contentType = http.DetectContentType(peek[:n])
		resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(peek[:n]), resp.Body))
	}
	if !strings.HasPrefix(contentType, "image/") {
		log.Println(contentType, "content type imagen no es image/")
		return fallbackUpload()
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err, "error ReadAll imagen")
		return fallbackUpload()
	}

	// Nombre y extensión
	ext := ".jpeg"
	if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
		ext = exts[0]
	}
	fileName := fmt.Sprintf("evidence_%d%s", time.Now().UnixNano(), ext)
	log.Println(fileName, "fileName imagen")

	// Subir a Supabase
	/*url, err := uploadToSupabase(ctx, fileName, data, contentType)
	if err != nil {
		log.Println(err, "error uploadToSupabase imagen")
		return fallbackUpload()
	}
	log.Println(url, "url fetchAndUploadSupportImage imagen")*/
	url := fmt.Sprintf("data:%s;base64,%s", "image/jpeg", base64.StdEncoding.EncodeToString(data))
	return url, nil
}

func fallbackUpload() (string, error) {
	// Ruta relativa: internal/worker/img/Comprobante-test.jpeg
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "internal", "worker", "img", "Comprobante-test.jpeg")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// uploadToSupabase(ctx, fmt.Sprintf("fallback_%d.jpeg", time.Now().UnixNano()), data, "image/jpeg")
	url := fmt.Sprintf("data:%s;base64,%s", "image/jpeg", base64.StdEncoding.EncodeToString(data))
	return url, err
}

/*func uploadToSupabase(ctx context.Context, name string, data []byte, contentType string) (string, error) {
	if config.C.SupabaseURLProject == "" || config.C.SupabaseBucket == "" || config.C.SupabaseAPIKey == "" {
		return "", fmt.Errorf("supabase env missing")
	}
	// Endpoint de Storage: POST /storage/v1/object/{bucket}/{path}
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", strings.TrimRight(config.C.SupabaseURLProject, "/"), config.C.SupabaseBucket, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+config.C.SupabaseAPIKey)
	req.Header.Set("apikey", config.C.SupabaseAPIKey)
	req.Header.Set("Content-Type", contentType)
	// Preferir URL pública
	req.Header.Set("x-upsert", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload failed: %s", string(body))
	}

	// Construir URL pública: {project}/storage/v1/object/public/{bucket}/{name}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", strings.TrimRight(config.C.SupabaseURLProject, "/"), config.C.SupabaseBucket, name)
	return publicURL, nil
}*/

func extractIdFromTransactionJSON(transactionJSON string) (string, error) {
	// Intenta ambos formatos
	if strings.Contains(transactionJSON, "\"id\":{") {
		idx := strings.Index(transactionJSON, "\"id\":{")
		start := strings.Index(transactionJSON[idx:], "\"$oid\":\"")
		if start == -1 {
			return "", fmt.Errorf("$oid not found in id")
		}
		start = idx + start + len("\"$oid\":\"")
		end := strings.Index(transactionJSON[start:], "\"")
		if end == -1 {
			return "", fmt.Errorf("malformed $oid")
		}
		return transactionJSON[start : start+end], nil
	}
	idx := strings.Index(transactionJSON, "\"id\":\"")
	if idx == -1 {
		return "", fmt.Errorf("id not found")
	}
	start := idx + len("\"id\":\"")
	end := strings.Index(transactionJSON[start:], "\"")
	if end == -1 {
		return "", fmt.Errorf("malformed id")
	}
	return transactionJSON[start : start+end], nil
}

func setMongoSupportURL(ctx context.Context, idHex string, uploadedURL string) error {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return err
	}
	_, err = db.Mongo().Collection("transactions").UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"support_url": uploadedURL, "updatedAt": time.Now()}},
	)
	return err
}

// extractMidFromWhatsAppURL extrae el parámetro mid de la URL de WhatsApp
/*func extractMidFromWhatsAppURL(whatsappURL string) (string, error) {
	log.Println(whatsappURL, "whatsappURL para extraer mid")
	// Buscar el parámetro mid en la URL
	idx := strings.Index(whatsappURL, "mid=")
	if idx == -1 {
		return "", fmt.Errorf("mid parameter not found in WhatsApp URL")
	}
	start := idx + len("mid=")
	end := strings.Index(whatsappURL[start:], "u0026source=getMedia")
	cleanURL := strings.Replace(whatsappURL[start:start+end], "\\", "", -1)
	if end == -1 {
		// Si no hay & después de mid, tomar hasta el final
		return cleanURL, nil
	}
	return cleanURL, nil
}*/

// getRealImageURLFromMeta obtiene la URL real de la imagen desde Graph API de Meta
func getRealImageURLFromMeta(ctx context.Context, mid string) (string, error) {
	// Construir URL de Graph API
	graphURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", mid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, graphURL, nil)

	if err != nil {
		return "[graph otro ERROR]", err
	}

	if config.C.BearerTokenMeta != "" {
		req.Header.Set("Authorization", "Bearer "+config.C.BearerTokenMeta)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "[graph.facebook ERROR]", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Graph API request failed with status: %d", resp.StatusCode)
	}

	// Parsear respuesta JSON
	var response struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if response.URL == "" {
		return "", fmt.Errorf("empty URL in Graph API response")
	}

	return response.URL, nil
}
