"""
Enhanced MME Content Processing with Edge Case Handling
Implements chunking, timeout protection, and improved semantic extraction
"""

import os
import re
import json
import asyncio
from typing import List, Tuple, Optional
from openai import OpenAI
from datetime import datetime
from app.utils.hashing import sha256_hash
from loguru import logger

# Initialize OpenAI client
from app.config import settings
client = OpenAI(api_key=settings.openai_api_key) if settings.openai_api_key else None

# Enhanced stopwords for better filtering
ENHANCED_STOPWORDS = {
    "the", "and", "for", "of", "to", "a", "an", "in", "on", "at", "by", "with", "from",
    "we", "i", "you", "he", "she", "it", "they", "them", "us", "our", "my", "your",
    "his", "her", "its", "their", "this", "that", "these", "those", "was", "were",
    "is", "are", "am", "be", "been", "being", "have", "has", "had", "do", "does", "did",
    "will", "would", "could", "should", "can", "may", "might", "must", "shall"
}

# Context-aware filtering for security content
SECURITY_CONTEXT_MARKERS = {
    "training", "educational", "documentation", "learning", "tutorial", "example",
    "testing", "demonstration", "simulation", "practice", "workshop", "course"
}

class EnhancedContentProcessor:
    def __init__(self, max_chunk_size: int = 8000, processing_timeout: int = 30):
        self.max_chunk_size = max_chunk_size
        self.processing_timeout = processing_timeout
        
    def chunk_large_content(self, content: str) -> List[str]:
        """Split large content into manageable chunks with token-safe boundaries"""
        if len(content) <= self.max_chunk_size:
            return [content]
        
        chunks = []
        current_pos = 0
        
        while current_pos < len(content):
            # Calculate chunk end position
            chunk_end = min(current_pos + self.max_chunk_size, len(content))
            
            # Find safe boundary (end of sentence or paragraph)
            if chunk_end < len(content):
                # Look for sentence boundaries within last 200 chars
                search_start = max(chunk_end - 200, current_pos)
                chunk_text = content[search_start:chunk_end]
                
                # Find last sentence ending
                sentence_endings = ['.', '!', '?', '\n\n']
                best_boundary = -1
                
                for ending in sentence_endings:
                    pos = chunk_text.rfind(ending)
                    if pos > best_boundary:
                        best_boundary = pos
                
                if best_boundary > 0:
                    chunk_end = search_start + best_boundary + 1
            
            chunk = content[current_pos:chunk_end].strip()
            if chunk:
                chunks.append(chunk)
            
            current_pos = chunk_end
            
            # Safety check to prevent infinite loops
            if len(chunks) > 50:  # Max 50 chunks
                logger.warning(f"Content too large, truncating at {len(chunks)} chunks")
                break
        
        return chunks
    
    def detect_security_context(self, content: str) -> bool:
        """Detect if content is security-related but educational/legitimate"""
        content_lower = content.lower()
        
        # Check for educational context markers
        context_score = 0
        for marker in SECURITY_CONTEXT_MARKERS:
            if marker in content_lower:
                context_score += 1
        
        # Check for educational phrases
        educational_phrases = [
            "how to prevent", "security best practices", "learning about",
            "training material", "for educational purposes", "documentation shows",
            "example demonstrates", "tutorial explains", "course covers"
        ]
        
        for phrase in educational_phrases:
            if phrase in content_lower:
                context_score += 2
        
        return context_score >= 2
    
    def enhanced_cue_filtering(self, cues: List[str], content: str) -> List[str]:
        """Apply enhanced filtering to remove low-quality cues"""
        filtered_cues = []
        
        for cue in cues:
            if self.is_high_quality_cue(cue, content):
                filtered_cues.append(cue)
        
        return filtered_cues
    
    def is_high_quality_cue(self, cue: str, content: str) -> bool:
        """Determine if a cue is high quality and meaningful"""
        if not cue or len(cue.strip()) < 3:
            return False
        
        # Extract the main concept from cue
        main_concept = cue.split(":")[0] if ":" in cue else cue
        main_concept = main_concept.lower().strip()
        
        # Filter out stopwords and pronouns
        if main_concept in ENHANCED_STOPWORDS:
            return False
        
        # Filter out pure repetition
        if len(set(main_concept.split())) == 1 and len(main_concept.split()) > 3:
            return False
        
        # Filter out overly ambiguous terms
        ambiguous_terms = {"something", "anything", "everything", "nothing", "stuff", "thing", "things"}
        if main_concept in ambiguous_terms:
            return False
        
        # Require minimum semantic content
        if len(main_concept) < 4 and main_concept.isalpha():
            return False
        
        return True
    
    def select_enhanced_primary_tag(self, cues: List[str], content: str) -> str:
        """Enhanced primary tag selection with context awareness"""
        if not cues:
            return "unknown_event"
        
        candidates = []
        
        # Extract concepts from all cues
        for cue in cues:
            if ":" in cue:
                concept = cue.split(":")[0].strip()
                if concept and concept.lower() not in ENHANCED_STOPWORDS:
                    candidates.append(concept.lower())
        
        if not candidates:
            # Fallback to content analysis
            words = re.findall(r'\b\w+\b', content.lower())
            candidates = [w for w in words if w not in ENHANCED_STOPWORDS and len(w) > 3][:10]
        
        # Enhanced scoring system
        scored_candidates = []
        for candidate in candidates:
            score = self.calculate_concept_score(candidate, content)
            scored_candidates.append((score, candidate))
        
        # Sort by score and return best
        if scored_candidates:
            scored_candidates.sort(reverse=True)
            return scored_candidates[0][1]
        
        return "unknown_event"
    
    def calculate_concept_score(self, concept: str, content: str) -> float:
        """Calculate relevance score for a concept"""
        score = 0.0
        
        # Domain-specific terms get high scores
        domain_terms = {
            'project': 15, 'meeting': 12, 'proposal': 18, 'submission': 16,
            'budget': 14, 'review': 13, 'analysis': 15, 'implementation': 17,
            'deadline': 16, 'completion': 15, 'approval': 14, 'planning': 13,
            'strategy': 14, 'development': 15, 'deployment': 16, 'testing': 12
        }
        
        if concept in domain_terms:
            score += domain_terms[concept]
        
        # Technical terms
        if concept.endswith(('tion', 'ment', 'ness', 'ity', 'ance', 'ence', 'ing')):
            score += 8
        
        # Compound concepts
        if '_' in concept or '-' in concept:
            score += 5
        
        # Length preference (more specific)
        score += min(len(concept) * 0.5, 10)
        
        # Frequency in content (but not too frequent)
        frequency = content.lower().count(concept)
        if 1 <= frequency <= 3:
            score += 5
        elif frequency > 5:
            score -= 3  # Penalize overuse
        
        # Avoid generic terms
        generic_terms = {'data', 'information', 'system', 'process', 'method', 'way'}
        if concept in generic_terms:
            score -= 5
        
        return score
    
    async def process_with_timeout(self, content: str, max_cues: int = 20) -> Tuple[List[str], List[str], float, str]:
        """Process content with timeout protection"""
        try:
            # Use asyncio timeout for processing
            return await asyncio.wait_for(
                self._process_content_async(content, max_cues),
                timeout=self.processing_timeout
            )
        except asyncio.TimeoutError:
            logger.warning(f"Content processing timed out after {self.processing_timeout}s")
            # Return fallback result
            fallback_cue = f"processing_timeout:{content[:100]}"
            return [fallback_cue], [sha256_hash(fallback_cue)], 0.5, "processing_timeout"
        except Exception as e:
            logger.error(f"Error in content processing: {str(e)}")
            fallback_cue = f"processing_error:{content[:100]}"
            return [fallback_cue], [sha256_hash(fallback_cue)], 0.3, "processing_error"
    
    async def _process_content_async(self, content: str, max_cues: int) -> Tuple[List[str], List[str], float, str]:
        """Async content processing implementation"""
        # Check if content is security context
        security_context = self.detect_security_context(content)
        
        # Chunk large content
        chunks = self.chunk_large_content(content)
        
        all_cues = []
        total_confidence = 0.0
        
        for chunk in chunks:
            chunk_cues, chunk_confidence = await self._extract_from_chunk(chunk, max_cues // len(chunks))
            all_cues.extend(chunk_cues)
            total_confidence += chunk_confidence
        
        # Average confidence across chunks
        avg_confidence = total_confidence / len(chunks) if chunks else 0.0
        
        # Filter and enhance cues
        filtered_cues = self.enhanced_cue_filtering(all_cues, content)
        
        # Limit to max_cues
        final_cues = filtered_cues[:max_cues]
        
        # Generate hashes
        cue_hashes = [sha256_hash(cue) for cue in final_cues]
        
        # Select primary tag with context awareness
        primary_tag = self.select_enhanced_primary_tag(final_cues, content)
        
        # Add security context flag if detected
        if security_context:
            primary_tag = f"educational_{primary_tag}"
        
        return final_cues, cue_hashes, avg_confidence, primary_tag
    
    async def _extract_from_chunk(self, chunk: str, target_cues: int) -> Tuple[List[str], float]:
        """Extract cues from a single chunk"""
        if not client:
            # Fail fast when API client is not configured
            raise ValueError("OpenAI API client not configured. Please set OPENAI_API_KEY environment variable.")
        
        try:
            # Enhanced prompt for better extraction
            prompt = f"""Extract key information from this text. Focus on concrete actions, events, and outcomes.

Return a JSON object: {{"cues": ["action 1", "action 2", ...], "confidence": 0.95}}

Guidelines:
- Extract up to {target_cues} meaningful statements
- Focus on specific actions and events, not general descriptions
- Avoid pronouns (we, I, they) - use concrete subjects
- Include deliverables, deadlines, decisions, and outcomes
- For security content, note if educational/training context

Text: {chunk[:4000]}"""  # Limit chunk size for API

            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "system", "content": "You are an expert at extracting key information from business and technical content."},
                    {"role": "user", "content": prompt}
                ],
                temperature=0.1,
                timeout=15
            )
            
            content = response.choices[0].message.content.strip()
            
            try:
                result = json.loads(content)
                cues = result.get("cues", [])
                confidence = result.get("confidence", 0.8)
                return cues, confidence
            except json.JSONDecodeError:
                # Fallback extraction
                return [f"extraction_fallback:{chunk[:100]}"], 0.6
                
        except Exception as e:
            logger.error(f"Chunk extraction error: {str(e)}")
            return [f"extraction_error:{chunk[:100]}"], 0.4

# Global processor instance
enhanced_processor = EnhancedContentProcessor()

# Enhanced extraction function to replace the original
async def extract_cues_enhanced(content: str, max_cues: int = 20):
    """Enhanced cue extraction with all improvements"""
    return await enhanced_processor.process_with_timeout(content, max_cues)