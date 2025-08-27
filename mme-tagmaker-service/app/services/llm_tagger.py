import os, hashlib, re, json
from openai import OpenAI
from datetime import datetime
from app.utils.hashing import sha256_hash
from app.models.tag import Tag
from app.services.domain_lexicon import get_domain_type, get_synonyms

def normalize_label(label: str) -> str:
    """Normalize tag label: lowercase, trim, collapse spaces"""
    if not label:
        return ""
    # Convert to lowercase, trim whitespace, collapse multiple spaces
    normalized = re.sub(r'\s+', ' ', label.strip().lower())
    return normalized
from loguru import logger

# Initialize OpenAI client with optional API key for testing
from app.config import settings
client = OpenAI(api_key=settings.openai_api_key) if settings.openai_api_key else None

# Expanded stopwords for primary tag filtering
STOPWORDS = {
    "the", "and", "for", "of", "to", "a", "an", "in", "on", "at", "by", "with", "from",
    "we", "i", "you", "he", "she", "it", "they", "them", "us", "our", "my", "your",
    "his", "her", "its", "their", "this", "that", "these", "those", "was", "were",
    "is", "are", "am", "be", "been", "being", "have", "has", "had", "do", "does", "did"
}

def extract_semantic_concepts(sentence: str) -> tuple[str, str]:
    """Extract meaningful concept and detail from sentence using semantic analysis"""
    words = re.findall(r'\b\w+\b', sentence.lower())
    
    # Look for nouns, verbs, and specific domain terms
    concept_candidates = []
    
    # Simple heuristics for concept extraction
    for i, word in enumerate(words):
        if word not in STOPWORDS and len(word) > 2:
            # Prefer nouns and domain-specific terms
            if word.endswith(('ion', 'ment', 'ness', 'ity', 'ance', 'ence')):  # Noun endings
                concept_candidates.append(word)
            elif word in ['proposal', 'submission', 'report', 'project', 'deadline', 'task', 'meeting', 'review']:
                concept_candidates.append(word)
            elif i > 0 and words[i-1] not in STOPWORDS:  # Context-dependent words
                concept_candidates.append(f"{words[i-1]}_{word}")
    
    # If no good concept found, use longest non-stopword
    if not concept_candidates:
        concept_candidates = [w for w in words if w not in STOPWORDS and len(w) > 2]
    
    concept = concept_candidates[0] if concept_candidates else "unknown_action"
    detail = " ".join(words[:8])  # First 8 words as detail
    
    return concept, detail

def make_cue(sentence: str) -> str:
    """Generate structured cue with semantic concept extraction"""
    concept, detail = extract_semantic_concepts(sentence)
    return f"{concept}:{detail}".strip(":")

def select_primary_tag(cues: list[str], content: str) -> str:
    """Select the most semantically meaningful primary tag from cues"""
    candidates = []
    
    # Extract concepts from all cues
    for cue in cues:
        if ":" in cue:
            concept = cue.split(":")[0]
            if concept not in STOPWORDS and len(concept) > 2:
                candidates.append(concept)
    
    if not candidates:
        # Fallback: extract from content directly
        words = re.findall(r'\b\w+\b', content.lower())
        candidates = [w for w in words if w not in STOPWORDS and len(w) > 3]
    
    # Score candidates by semantic value
    scored_candidates = []
    for candidate in candidates:
        score = 0
        # Prefer domain terms
        if candidate in ['proposal', 'submission', 'project', 'deadline', 'task', 'meeting', 'review', 'report']:
            score += 10
        # Prefer nouns (rough heuristic)
        if candidate.endswith(('ion', 'ment', 'ness', 'ity', 'ance', 'ence', 'al', 'ing')):
            score += 5
        # Prefer compound concepts
        if "_" in candidate:
            score += 3
        # Prefer longer terms (more specific)
        score += len(candidate)
        
        scored_candidates.append((score, candidate))
    
    # Return highest scoring candidate
    if scored_candidates:
        scored_candidates.sort(reverse=True)
        return scored_candidates[0][1]
    
    return "unknown_event"

def determine_tag_type(label: str) -> str:
    """Determine tag type based on label content"""
    label_lower = label.lower()
    
    # Action words
    action_words = [
        "submit", "create", "build", "implement", "deploy", "test", "review",
        "complete", "finish", "start", "begin", "launch", "release", "update",
        "fix", "resolve", "solve", "process", "handle", "manage", "execute",
        "run", "perform", "conduct", "carry", "out", "deliver", "provide"
    ]
    
    # Object words
    object_words = [
        "form", "document", "file", "report", "proposal", "budget", "plan",
        "system", "application", "database", "api", "service", "module",
        "component", "interface", "model", "framework", "library", "tool",
        "platform", "environment", "configuration", "setting", "parameter"
    ]
    
    # Error words
    error_words = [
        "error", "fail", "failure", "exception", "bug", "issue", "problem",
        "crash", "timeout", "invalid", "missing", "broken", "corrupt",
        "unavailable", "denied", "rejected", "cancelled", "aborted"
    ]
    
    # Status words
    status_words = [
        "status", "state", "condition", "ready", "pending", "active",
        "inactive", "enabled", "disabled", "running", "stopped", "completed",
        "failed", "success", "approved", "rejected", "draft", "final",
        "published", "archived", "expired", "valid", "invalid"
    ]
    
    for word in action_words:
        if word in label_lower:
            return "action"
    
    for word in object_words:
        if word in label_lower:
            return "object"
    
    for word in error_words:
        if word in label_lower:
            return "error"
    
    for word in status_words:
        if word in label_lower:
            return "status"
    
    return "concept"

def determine_section(label: str, content: str) -> str:
    """Determine section based on label and content"""
    label_lower = label.lower()
    content_lower = content.lower()
    
    # Section keywords
    sections = {
        "funding-proposal": ["funding", "proposal", "grant", "budget", "cost", "financial"],
        "technical-implementation": ["technical", "implementation", "code", "development", "programming"],
        "project-management": ["project", "management", "timeline", "deadline", "milestone"],
        "research-analysis": ["research", "analysis", "study", "investigation", "examination"],
        "documentation": ["documentation", "document", "manual", "guide", "specification"]
    }
    
    for section, keywords in sections.items():
        for keyword in keywords:
            if keyword in label_lower or keyword in content_lower:
                return section
    
    return "general"

def extract_cues(content: str, max_cues: int = 20):
    # Fail fast if OpenAI API key is not configured
    if not client:
        logger.error("OpenAI API key not configured - LLM tagging service requires valid API key")
        raise ValueError("OpenAI API key not configured. Please set OPENAI_API_KEY environment variable.")
    
    # Validate content is not empty or whitespace
    if not content or not content.strip():
        logger.warning("Empty or whitespace-only content provided for tag extraction")
        raise ValueError("Content cannot be empty or contain only whitespace")
    
    # Check content length and truncate if too long
    content = content.strip()
    if len(content) > 8000:  # Limit to ~8k chars to prevent token overflow
        logger.warning(f"Content too long ({len(content)} chars), truncating to 8000 chars")
        content = content[:8000] + "..."
    
    # Enhanced prompt for better semantic extraction
    prompt = f"""Analyze the following text and extract key information. Focus on identifying the main actions, events, or concepts.

Return ONLY a valid JSON object with this exact structure:
{{"cues": ["action or event description 1", "action or event description 2", ...], "confidence": 0.95}}

Guidelines for extraction:
- Extract {max_cues} or fewer meaningful statements
- Focus on actions, decisions, completions, and outcomes
- Avoid pronouns (we, I, they) in favor of concrete actions
- Prefer specific events over general statements
- Include deadlines, submissions, reviews, and deliverables

Text to analyze:"""
    
    try:
        resp = client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[
                {"role": "system", "content": prompt},
                {"role": "user", "content": content}
            ],
            temperature=0.1,  # Lower temperature for more deterministic output
        )
        
        raw = resp.choices[0].message.content.strip()
        
        # Parse JSON response
        result = json.loads(raw)
        sentences = result.get("cues", [])
        confidence = result.get("confidence", 0.95)
        
        if not sentences:
            raise ValueError("LLM returned no cues from content analysis")
            
        # Process sentences into structured tags
        tags = []
        now = datetime.now()
        
        for sentence in sentences[:max_cues]:
            if not sentence.strip():
                continue
                
            concept, detail = extract_semantic_concepts(sentence)
            if concept and concept != "unknown_action":
                # Create structured tag with normalized label and domain typing
                normalized_label = normalize_label(concept)
                domain_type = get_domain_type(normalized_label)
                
                tag = Tag(
                    label=normalized_label,
                    section=determine_section(concept, content),
                    origin="agent",
                    scope="shared",
                    type=domain_type,
                    confidence=confidence,
                    links=[detail] if detail else [],
                    usageCount=1,
                    lastUsed=now
                )
                tags.append(tag)
        
        # Select primary tag
        primary_tag = select_primary_tag([f"{tag.label}:{tag.links[0] if tag.links else ''}" for tag in tags], content)
        
        return tags, confidence, primary_tag
        
    except Exception as e:
        logger.error(f"Error in LLM extraction: {str(e)}")
        raise ValueError(f"LLM extraction failed: {str(e)}")
