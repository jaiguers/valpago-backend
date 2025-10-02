use mongodb::{Client, Database};

pub async fn create_mongo_client() -> anyhow::Result<Database> {
    let uri = std::env::var("MONGODB_URI").expect("MONGODB_URI is required");
    let db_name = std::env::var("MONGODB_DB").unwrap_or_else(|_| "valpago".into());
    let client = Client::with_uri_str(uri).await?;
    Ok(client.database(&db_name))
}

