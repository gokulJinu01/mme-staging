"""
MSE + MME Conflict Resolution System
Handles conflicts between security detection and memory tagging
"""

import re
from typing import Dict, Any, Tuple
from loguru import logger

class SecurityMemoryConflictResolver:
    def __init__(self):
        # Educational context indicators
        self.educational_indicators = {
            "training", "educational", "documentation", "learning", "tutorial", 
            "example", "testing", "demonstration", "simulation", "practice",
            "workshop", "course", "guide", "manual", "reference", "how-to"
        }
        
        # Legitimate security work indicators
        self.security_work_indicators = {
            "security audit", "penetration testing", "vulnerability assessment",
            "security training", "security documentation", "security review",
            "threat modeling", "security implementation", "security testing",
            "cybersecurity", "infosec", "security awareness", "compliance audit"
        }
        
        # High-risk patterns that should always be blocked
        self.always_block_patterns = [
            r"(?i)drop\s+table",
            r"(?i)delete\s+from\s+\*",
            r"(?i)rm\s+-rf\s+/",
            r"(?i)format\s+c:",
            r"(?i)exec\s*\(\s*['\"].*system.*['\"]",
        ]
    
    def resolve_conflict(self, mse_result: Dict[str, Any], mme_result: Dict[str, Any], 
                        original_content: str) -> Dict[str, Any]:
        """
        Resolve conflicts between MSE security detection and MME memory tagging
        
        Returns enhanced result with conflict resolution
        """
        
        # Extract key metrics
        mse_score = mse_result.get("score", 0)
        mse_action = mse_result.get("action", "allow")
        mse_rules = mse_result.get("matched_rules", [])
        
        mme_primary_tag = mme_result.get("primary_tag", "")
        mme_confidence = mme_result.get("confidence", 0)
        mme_cues = mme_result.get("cues", [])
        
        # Analyze context
        context_analysis = self._analyze_content_context(original_content)
        
        # Determine if this is a false positive scenario
        is_educational = context_analysis["is_educational"]
        is_security_work = context_analysis["is_security_work"]
        has_critical_threats = self._has_critical_threats(original_content)
        
        # Build resolution result
        resolution = {
            "original_mse_result": mse_result,
            "original_mme_result": mme_result,
            "context_analysis": context_analysis,
            "conflict_detected": False,
            "resolution_applied": False,
            "final_action": mse_action,
            "final_score": mse_score,
            "enhanced_primary_tag": mme_primary_tag,
            "allow_context": False,
            "conflict_reason": None
        }
        
        # Detect conflict
        if mse_score > 50 and mme_confidence > 0.8:
            resolution["conflict_detected"] = True
            resolution["conflict_reason"] = f"MSE blocked (score: {mse_score}) vs MME high confidence ({mme_confidence:.2f})"
            
            # Apply resolution logic
            if has_critical_threats:
                # Always block critical threats regardless of context
                resolution["final_action"] = "block"
                resolution["final_score"] = mse_score
                resolution["resolution_applied"] = True
                resolution["conflict_reason"] += " - Critical threat detected, blocking"
                
            elif is_educational or is_security_work:
                # Reduce scoring for educational/legitimate security content
                adjusted_score = max(0, mse_score - 40)
                
                if adjusted_score < 30:
                    resolution["final_action"] = "allow"
                    resolution["final_score"] = adjusted_score
                    resolution["allow_context"] = True
                    resolution["enhanced_primary_tag"] = f"security_{mme_primary_tag}"
                    resolution["resolution_applied"] = True
                    resolution["conflict_reason"] += f" - Educational context detected, score reduced to {adjusted_score}"
                else:
                    resolution["final_action"] = "flag"
                    resolution["final_score"] = adjusted_score
                    resolution["enhanced_primary_tag"] = f"flagged_security_{mme_primary_tag}"
                    resolution["resolution_applied"] = True
                    resolution["conflict_reason"] += f" - Educational context with elevated score {adjusted_score}"
            
            else:
                # Unknown context - apply moderate adjustment
                adjusted_score = max(0, mse_score - 15)
                resolution["final_score"] = adjusted_score
                
                if adjusted_score >= 70:
                    resolution["final_action"] = "block"
                elif adjusted_score >= 40:
                    resolution["final_action"] = "flag"
                else:
                    resolution["final_action"] = "allow"
                
                resolution["resolution_applied"] = True
                resolution["conflict_reason"] += f" - Moderate adjustment applied, score: {adjusted_score}"
        
        # Enhance memory tagging based on security context
        if mse_score > 30:
            # Add security context to memory tags
            security_aware_tag = self._create_security_aware_tag(mme_primary_tag, mse_rules, context_analysis)
            resolution["enhanced_primary_tag"] = security_aware_tag
        
        return resolution
    
    def _analyze_content_context(self, content: str) -> Dict[str, Any]:
        """Analyze content to determine context and legitimacy"""
        content_lower = content.lower()
        
        # Educational context detection
        educational_score = 0
        for indicator in self.educational_indicators:
            if indicator in content_lower:
                educational_score += 1
        
        # Security work context detection
        security_work_score = 0
        for indicator in self.security_work_indicators:
            if indicator in content_lower:
                security_work_score += 2
        
        # Check for educational phrases
        educational_phrases = [
            "for learning purposes", "training material", "educational example",
            "documentation shows", "tutorial demonstrates", "course explains",
            "best practices", "security awareness", "how to prevent",
            "security guidelines", "recommended approach"
        ]
        
        for phrase in educational_phrases:
            if phrase in content_lower:
                educational_score += 3
        
        # Check for legitimate security work
        legitimate_work_phrases = [
            "security assessment", "penetration test", "vulnerability scan",
            "security audit", "compliance check", "security implementation",
            "security monitoring", "incident response", "threat analysis"
        ]
        
        for phrase in legitimate_work_phrases:
            if phrase in content_lower:
                security_work_score += 3
        
        return {
            "is_educational": educational_score >= 3,
            "is_security_work": security_work_score >= 4,
            "educational_score": educational_score,
            "security_work_score": security_work_score,
            "content_length": len(content),
            "likely_legitimate": educational_score >= 2 or security_work_score >= 3
        }
    
    def _has_critical_threats(self, content: str) -> bool:
        """Check if content contains critical threats that should always be blocked"""
        for pattern in self.always_block_patterns:
            if re.search(pattern, content):
                return True
        
        # Check for destructive commands
        destructive_indicators = [
            "rm -rf /", "format c:", "del /f /s /q", "DROP DATABASE",
            "TRUNCATE TABLE", "DELETE FROM users", "shutdown -h now",
            ":(){ :|:& };:", "while true; do", "fork bomb"
        ]
        
        content_lower = content.lower()
        for indicator in destructive_indicators:
            if indicator.lower() in content_lower:
                return True
        
        return False
    
    def _create_security_aware_tag(self, primary_tag: str, matched_rules: list, 
                                  context_analysis: Dict[str, Any]) -> str:
        """Create security-aware primary tag"""
        
        if context_analysis["is_educational"]:
            return f"educational_security_{primary_tag}"
        elif context_analysis["is_security_work"]:
            return f"legitimate_security_{primary_tag}"
        elif matched_rules:
            # Extract rule types for context
            rule_types = set()
            for rule in matched_rules:
                rule_id = rule.get("ruleId", "")
                if "SQL" in rule_id:
                    rule_types.add("sql")
                elif "PROMPT" in rule_id or "INJ" in rule_id:
                    rule_types.add("prompt")
                elif "MAL" in rule_id or "CMD" in rule_id:
                    rule_types.add("malware")
            
            if rule_types:
                security_context = "_".join(sorted(rule_types))
                return f"security_{security_context}_{primary_tag}"
        
        return f"security_flagged_{primary_tag}"

# Global resolver instance
conflict_resolver = SecurityMemoryConflictResolver()

def resolve_mse_mme_conflict(mse_result: Dict[str, Any], mme_result: Dict[str, Any], 
                           content: str) -> Dict[str, Any]:
    """
    Public interface for conflict resolution
    """
    return conflict_resolver.resolve_conflict(mse_result, mme_result, content)