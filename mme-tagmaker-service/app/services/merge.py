from datetime import datetime
from typing import List, Dict

def build_delta(tag: str,
                cues: List[str],
                cue_hashes: List[str],
                related: Dict[str,int]) -> Dict:
    now = datetime.utcnow().isoformat() + "Z"
    ops = {
        "$inc":  {"metrics.useCount": 1},
        "$set":  {"metrics.lastUsedAt": now},
        "$addToSet": {
            "context.cues": {"$each": cues},
            "context.cueHashes": {"$each": cue_hashes}
        },
        "$inc_related": related  # custom operator, handled downstream
    }
    return {"tag": tag, "ops": ops}
