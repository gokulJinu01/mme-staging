import hashlib

def sha256_hash(text: str) -> str:
    """Generate SHA256 hash of input text"""
    return hashlib.sha256(text.encode('utf-8')).hexdigest()
