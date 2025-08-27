"""
Domain Lexicon for Finance and Compliance
Provides controlled vocabulary and type mapping for semantic understanding
"""

DOMAIN_LEXICON = {
    "finance": {
        "budget": {
            "synonyms": ["QBR", "quarterly forecast", "budget Q1", "budget Q2", "budget Q3", "budget Q4", "financial plan", "spending plan"],
            "type": "finance.budget",
            "subtypes": {
                "quarterly": ["Q1", "Q2", "Q3", "Q4", "quarterly", "quarter"],
                "annual": ["annual", "yearly", "fiscal year", "FY"],
                "operational": ["opex", "operational expense", "operating budget"]
            }
        },
        "forecast": {
            "synonyms": ["projection", "estimate", "prediction", "outlook"],
            "type": "finance.forecast"
        },
        "compliance": {
            "synonyms": ["policy", "SOC2", "PCI", "ISO27001", "audit", "regulation", "governance"],
            "type": "compliance.policy",
            "subtypes": {
                "security": ["SOC2", "ISO27001", "security audit", "cybersecurity"],
                "financial": ["PCI", "financial audit", "SOX"],
                "operational": ["operational audit", "process compliance"]
            }
        }
    },
    "compliance": {
        "policy": {
            "synonyms": ["procedure", "guideline", "standard", "requirement", "mandate"],
            "type": "compliance.policy"
        },
        "audit": {
            "synonyms": ["review", "assessment", "evaluation", "examination", "inspection"],
            "type": "compliance.audit"
        },
        "regulation": {
            "synonyms": ["law", "statute", "requirement", "mandate", "directive"],
            "type": "compliance.regulation"
        }
    },
    "project": {
        "deadline": {
            "synonyms": ["due date", "target date", "milestone", "deliverable date"],
            "type": "project.deadline"
        },
        "submission": {
            "synonyms": ["delivery", "handover", "turnover", "completion"],
            "type": "project.submission"
        }
    }
}

def get_domain_type(label: str) -> str:
    """Determine domain type from label using lexicon"""
    label_lower = label.lower()
    
    # Check each domain and category
    for domain, categories in DOMAIN_LEXICON.items():
        for category, info in categories.items():
            # Check main category
            if category in label_lower:
                return info.get("type", f"{domain}.{category}")
            
            # Check synonyms
            synonyms = info.get("synonyms", [])
            for synonym in synonyms:
                if synonym.lower() in label_lower:
                    return info.get("type", f"{domain}.{category}")
            
            # Check subtypes
            subtypes = info.get("subtypes", {})
            for subtype, subtype_synonyms in subtypes.items():
                for subtype_synonym in subtype_synonyms:
                    if subtype_synonym.lower() in label_lower:
                        return f"{info.get('type', f'{domain}.{category}')}.{subtype}"
    
    return "general.concept"

def get_synonyms(label: str) -> list:
    """Get synonyms for a given label"""
    label_lower = label.lower()
    
    for domain, categories in DOMAIN_LEXICON.items():
        for category, info in categories.items():
            if category in label_lower:
                return info.get("synonyms", [])
            
            synonyms = info.get("synonyms", [])
            for synonym in synonyms:
                if synonym.lower() in label_lower:
                    return info.get("synonyms", [])
    
    return []
