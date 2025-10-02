use argon2::{password_hash::{PasswordHash, PasswordHasher, PasswordVerifier, SaltString}, Argon2};
use rand::rngs::OsRng;

pub fn hash_password(plain: &str) -> anyhow::Result<String> {
    let salt = SaltString::generate(&mut OsRng);
    let argon2 = Argon2::default();
    let hash = argon2.hash_password(plain.as_bytes(), &salt)?.to_string();
    Ok(hash)
}

pub fn verify_password(hash: &str, plain: &str) -> bool {
    let parsed = PasswordHash::new(hash);
    match parsed {
        Ok(ph) => Argon2::default().verify_password(plain.as_bytes(), &ph).is_ok(),
        Err(_) => false,
    }
}

