use mongodb::{bson::doc, Database};

pub async fn validate_api_key(db: &Database, key: &str) -> anyhow::Result<bool> {
    let coll = db.collection::<mongodb::bson::Document>("api_keys");
    let filter = doc! { "key": key, "active": true };
    let found = coll.find_one(filter, None).await?;
    Ok(found.is_some())
}

