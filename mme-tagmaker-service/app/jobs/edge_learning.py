"""
Edge Learning Job

Implements EMA-based edge weight updates from pack acceptance/rejection events.
Updates tag_edges collection with new weights and prunes to top-M per tag.
"""

import os
import time
import logging
from typing import List, Dict, Tuple, Optional
from datetime import datetime, timedelta
import pymongo
from prometheus_client import Counter, Histogram

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Prometheus metrics
EDGES_UPDATED_TOTAL = Counter(
    'mme_edges_updated_total',
    'Total number of edges updated by edge learning',
    ['org']
)

EDGE_LEARNING_DURATION = Histogram(
    'mme_edge_learning_duration_seconds',
    'Duration of edge learning operations',
    ['org']
)

class EdgeLearningJob:
    """Edge learning job that updates tag edge weights based on pack events."""
    
    def __init__(self):
        # Configuration from environment variables
        self.eta = float(os.getenv('MME_LEARN_ETA', '0.05'))  # Learning rate
        self.r = float(os.getenv('MME_LEARN_R', '0.03'))      # Reward for acceptance
        self.d = float(os.getenv('MME_LEARN_D', '0.01'))      # Penalty for rejection
        self.w_max = float(os.getenv('MME_LEARN_WMAX', '1.0'))  # Maximum weight
        self.window_hours = int(os.getenv('MME_LEARN_WINDOW_HOURS', '24'))  # Event window
        self.max_edges_per_tag = int(os.getenv('MME_MAX_EDGES_PER_TAG', '32'))  # Top-M
        
        # MongoDB connection
        self.mongo_client = pymongo.MongoClient(os.getenv('MONGODB_URI', 'mongodb://localhost:27017/'))
        self.db = self.mongo_client['mme']
        self.pack_events_collection = self.db['pack_events']
        self.tag_edges_collection = self.db['tag_edges']
        
        logger.info(f"EdgeLearningJob initialized with eta={self.eta}, r={self.r}, d={self.d}, w_max={self.w_max}")
    
    def _get_recent_pack_events(self, org_id: str, hours: Optional[int] = None) -> List[Dict]:
        """Get recent pack events for an organization within the specified window."""
        if hours is None:
            hours = self.window_hours
            
        cutoff_time = int((datetime.now() - timedelta(hours=hours)).timestamp())
        
        events = list(self.pack_events_collection.find({
            'orgId': org_id,
            'ts': {'$gte': cutoff_time}
        }))
        
        logger.info(f"Found {len(events)} pack events for org {org_id} in last {hours} hours")
        return events
    
    def _build_tag_pairs(self, tags: List[str]) -> List[Tuple[str, str]]:
        """Build unique unordered tag pairs from a list of tags."""
        pairs = []
        for i in range(len(tags)):
            for j in range(i + 1, len(tags)):
                pairs.append((tags[i], tags[j]))
        return pairs
    
    def _clip_weight(self, weight: float) -> float:
        """Clip weight to [0, w_max] range."""
        return max(0.0, min(weight, self.w_max))
    
    def _update_edge_weight(self, current_weight: float, accepted: bool) -> float:
        """Update edge weight using EMA formula."""
        # Calculate reward/penalty
        reward = self.r if accepted else -self.d
        
        # EMA update formula
        new_weight = current_weight + reward
        clipped_weight = self._clip_weight(new_weight)
        
        # EMA smoothing
        updated_weight = (1 - self.eta) * current_weight + self.eta * clipped_weight
        
        return self._clip_weight(updated_weight)
    
    def _get_current_edges(self, org_id: str, tag: str) -> Dict[str, float]:
        """Get current edges for a tag from the database."""
        doc = self.tag_edges_collection.find_one({'orgId': org_id, 'tag': tag})
        if not doc:
            return {}
        
        edges = {}
        for edge in doc.get('edges', []):
            edges[edge['to']] = edge['w']
        
        return edges
    
    def _upsert_edges(self, org_id: str, tag: str, edges: Dict[str, float]):
        """Upsert edges for a tag, keeping only top-M by weight."""
        # Sort edges by weight (descending) and take top-M
        sorted_edges = sorted(edges.items(), key=lambda x: x[1], reverse=True)
        top_edges = sorted_edges[:self.max_edges_per_tag]
        
        # Convert to edge format
        edge_list = []
        current_time = int(time.time())
        for to_tag, weight in top_edges:
            edge_list.append({
                'to': to_tag,
                'w': weight,
                'facetSrc': '',
                'facetDst': '',
                'ts': current_time
            })
        
        # Upsert document
        self.tag_edges_collection.update_one(
            {'orgId': org_id, 'tag': tag},
            {
                '$set': {
                    'orgId': org_id,
                    'tag': tag,
                    'edges': edge_list
                }
            },
            upsert=True
        )
        
        return len(edge_list)
    
    def run(self, org_id: Optional[str] = None, hours: Optional[int] = None) -> Dict[str, int]:
        """Run edge learning for all orgs or a specific org."""
        start_time = time.time()
        
        if org_id:
            orgs = [org_id]
        else:
            # Get all unique orgs from pack events
            orgs = self.pack_events_collection.distinct('orgId')
        
        total_updated = 0
        total_pruned = 0
        
        for org in orgs:
            org_start_time = time.time()
            
            try:
                # Get recent pack events
                events = self._get_recent_pack_events(org, hours)
                if not events:
                    continue
                
                # Track edge updates for this org
                org_edges_updated = 0
                org_edges_pruned = 0
                
                # Process each event
                for event in events:
                    tags = event.get('tags', [])
                    accepted = event.get('accepted', False)
                    
                    if len(tags) < 2:
                        continue
                    
                    # Build tag pairs
                    pairs = self._build_tag_pairs(tags)
                    
                    # Update edges for each pair
                    for tag1, tag2 in pairs:
                        # Update edge tag1 -> tag2
                        current_edges = self._get_current_edges(org, tag1)
                        current_weight = current_edges.get(tag2, 0.0)
                        new_weight = self._update_edge_weight(current_weight, accepted)
                        current_edges[tag2] = new_weight
                        
                        # Upsert and count
                        edges_count = self._upsert_edges(org, tag1, current_edges)
                        org_edges_updated += 1
                        
                        # Update edge tag2 -> tag1 (bidirectional)
                        current_edges2 = self._get_current_edges(org, tag2)
                        current_weight2 = current_edges2.get(tag1, 0.0)
                        new_weight2 = self._update_edge_weight(current_weight2, accepted)
                        current_edges2[tag1] = new_weight2
                        
                        # Upsert and count
                        edges_count2 = self._upsert_edges(org, tag2, current_edges2)
                        org_edges_updated += 1
                
                # Log results for this org
                org_duration = time.time() - org_start_time
                logger.info(f"Org {org}: Updated {org_edges_updated} edges in {org_duration:.2f}s")
                
                # Update metrics
                EDGES_UPDATED_TOTAL.labels(org=org).inc(org_edges_updated)
                EDGE_LEARNING_DURATION.labels(org=org).observe(org_duration)
                
                total_updated += org_edges_updated
                
            except Exception as e:
                logger.error(f"Error processing org {org}: {e}")
                continue
        
        total_duration = time.time() - start_time
        logger.info(f"Edge learning completed: {total_updated} edges updated in {total_duration:.2f}s")
        
        return {
            'updated': total_updated,
            'pruned': total_pruned,
            'duration_seconds': total_duration
        }

# Global instance
edge_learning_job = EdgeLearningJob()

def run_edge_learning():
    """Entry point for scheduled edge learning."""
    return edge_learning_job.run()

def run_edge_learning_for_org(org_id: str, hours: Optional[int] = None):
    """Entry point for manual edge learning for a specific org."""
    return edge_learning_job.run(org_id=org_id, hours=hours)
