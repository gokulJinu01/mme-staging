from pydantic import BaseModel, Field
from typing import Optional, List

class TagRequest(BaseModel):
    content: str = Field(..., description="Raw agent output")
    userId: str
    orgId: Optional[str] = "test-org"
    sessionId: Optional[str] = None
    source: str = "agent_output"
