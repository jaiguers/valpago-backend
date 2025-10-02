use mongodb::Database;
use redis::Client as RedisClient;
use tokio::task::JoinHandle;

use crate::realtime::notifier::Notifier;

pub fn spawn_redis_worker(mongodb: Database, redis: RedisClient, notifier: Notifier) -> JoinHandle<()> {
    tokio::spawn(async move {
        if let Err(e) = run(redis, notifier).await {
            tracing::error!(error = %e, "worker stopped");
        }
    })
}

async fn run(redis: RedisClient, notifier: Notifier) -> anyhow::Result<()> {
    use redis::AsyncCommands;
    let mut conn = redis.get_async_connection().await?;
    let stream = crate::db::redis::stream_namespace();
    let group = crate::db::redis::consumer_group();
    let name = crate::db::redis::consumer_name();

    // Ensure group exists
    let _ : Result<String, _> = redis::cmd("XGROUP")
        .arg("CREATE")
        .arg(&stream)
        .arg(&group)
        .arg("0")
        .arg("MKSTREAM")
        .query_async(&mut conn)
        .await;

    loop {
        let reply: redis::Value = redis::cmd("XREADGROUP")
            .arg("GROUP").arg(&group).arg(&name)
            .arg("BLOCK").arg(5000)
            .arg("COUNT").arg(100)
            .arg("STREAMS").arg(&stream).arg(">")
            .query_async(&mut conn)
            .await?;

        let messages = parse_streams(reply);
        for (id, payload) in messages {
            notifier.broadcast(&payload);
            // Acknowledge message
            let _: () = redis::cmd("XACK").arg(&stream).arg(&group).arg(&id).query_async(&mut conn).await?;
        }
    }
}

fn parse_streams(value: redis::Value) -> Vec<(String, String)> {
    // minimal parser for XREADGROUP reply
    match value {
        redis::Value::Bulk(streams) if !streams.is_empty() => {
            let mut out = Vec::new();
            // streams: [[stream_key, [[id, [k,v,k,v...]], ...]]]
            for s in streams {
                if let redis::Value::Bulk(entries) = s {
                    if entries.len() >= 2 {
                        if let redis::Value::Bulk(items) = &entries[1] {
                            for item in items {
                                if let redis::Value::Bulk(parts) = item {
                                    if parts.len() == 2 {
                                        if let (redis::Value::Data(id), redis::Value::Bulk(kvs)) = (&parts[0], &parts[1]) {
                                            let mut payload = None;
                                            let mut it = kvs.iter();
                                            while let (Some(k), Some(v)) = (it.next(), it.next()) {
                                                if let (redis::Value::Data(kb), redis::Value::Data(vb)) = (k, v) {
                                                    if kb == b"payload" { payload = Some(String::from_utf8_lossy(vb).to_string()); }
                                                }
                                            }
                                            if let Some(p) = payload { out.push((String::from_utf8_lossy(id).to_string(), p)); }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            out
        }
        _ => Vec::new(),
    }
}

