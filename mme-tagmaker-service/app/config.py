import os
from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # OpenAI Configuration
    openai_api_key: Optional[str] = None
    
    # Tagging Service Configuration (Optional - for tag generation only)
    tagging_service_url: Optional[str] = os.getenv("MME_TAGGING_SERVICE_URL")
    tagmaker_jwt_secret: Optional[str] = os.getenv("TAGMAKER_JWT_SECRET")
    enable_tagging_service: bool = os.getenv("ENABLE_TAGGING_SERVICE", "false").lower() == "true"
    
    # MongoDB Configuration for direct database access (Primary dependency)
    mongodb_uri: str = os.getenv("MONGODB_URI")
    mongodb_database: str = os.getenv("MONGODB_DATABASE", "mme")
    mongodb_collection: str = os.getenv("MONGODB_COLLECTION", "memories")
    
    # Server Configuration
    host: str = "0.0.0.0"
    port: int = 8000
    
    class Config:
        env_file = ".env"
        case_sensitive = False

settings = Settings()
