"""
Pytest configuration for MME Tagmaker Service integration tests.

Provides real MongoDB container and test fixtures for edge learning tests.
"""

import os
import time
from typing import List, Dict, Any
import pytest
from testcontainers.mongodb import MongoDbContainer
import pymongo
from pymongo import MongoClient


@pytest.fixture(scope="session")
def mongo_container():
    """Start MongoDB container for the test session."""
    with MongoDbContainer("mongo:7.0") as container:
        # Wait for MongoDB to be ready
        time.sleep(5)
        yield container


@pytest.fixture(scope="session")
def mongo_uri(mongo_container):
    """Get MongoDB URI for the test container."""
    return mongo_container.get_connection_url()


@pytest.fixture
def mongo_client(mongo_uri):
    """Get MongoDB client for tests."""
    client = MongoClient(mongo_uri)
    yield client
    # Clean up after each test
    client.drop_database("mme")
    client.close()


@pytest.fixture
def mongo_db(mongo_client):
    """Get MongoDB database for tests."""
    db = mongo_client["mme"]
    
    # Ensure indexes are created
    ensure_indexes(db)
    
    return db


def ensure_indexes(db):
    """Create all required indexes for testing."""
    # Tag edges collection indexes
    tag_edges = db["tag_edges"]
    
    # Unique index for orgId + tag
    tag_edges.create_index(
        [("orgId", pymongo.ASCENDING), ("tag", pymongo.ASCENDING)],
        unique=True,
        name="tag_edges_org_tag_unique"
    )
    
    # Index for orgId + edges.to
    tag_edges.create_index(
        [("orgId", pymongo.ASCENDING), ("edges.to", pymongo.ASCENDING)],
        name="tag_edges_org_edges_to"
    )
    
    # Pack events collection indexes
    pack_events = db["pack_events"]
    
    # Index for orgId + timestamp
    pack_events.create_index(
        [("orgId", pymongo.ASCENDING), ("ts", pymongo.DESCENDING)],
        name="pack_events_org_ts"
    )
    
    # Index for orgId + packId
    pack_events.create_index(
        [("orgId", pymongo.ASCENDING), ("packId", pymongo.ASCENDING)],
        name="pack_events_org_pack"
    )


def seed_pack_events(mongo_db, org_id: str, events: List[Dict[str, Any]]):
    """Seed pack events for testing."""
    collection = mongo_db["pack_events"]
    
    for event in events:
        # Ensure required fields
        event.setdefault("orgId", org_id)
        event.setdefault("ts", int(time.time()))
        
        collection.insert_one(event)


def seed_tag_edges(mongo_db, org_id: str, edges: Dict[str, List[Dict[str, Any]]]):
    """Seed tag edges for testing."""
    collection = mongo_db["tag_edges"]
    
    for tag, tag_edges in edges.items():
        # Convert edges to documents
        edge_docs = []
        for edge in tag_edges:
            edge_docs.append({
                "to": edge["to"],
                "w": edge["w"],
                "facetSrc": edge.get("facetSrc", ""),
                "facetDst": edge.get("facetDst", ""),
                "ts": edge.get("ts", int(time.time())),
            })
        
        doc = {
            "orgId": org_id,
            "tag": tag,
            "edges": edge_docs,
        }
        
        collection.insert_one(doc)


def get_tag_edges(mongo_db, org_id: str, tag: str) -> List[Dict[str, Any]]:
    """Get tag edges for verification."""
    collection = mongo_db["tag_edges"]
    
    doc = collection.find_one({"orgId": org_id, "tag": tag})
    if not doc:
        return []
    
    return doc.get("edges", [])


def get_pack_events(mongo_db, org_id: str) -> List[Dict[str, Any]]:
    """Get pack events for verification."""
    collection = mongo_db["pack_events"]
    
    cursor = collection.find({"orgId": org_id})
    return list(cursor)


@pytest.fixture
def sample_pack_events():
    """Sample pack events for testing."""
    current_time = int(time.time())
    return [
        {
            "orgId": "test-org",
            "projectId": "test-project",
            "packId": "pack-1",
            "accepted": True,
            "tags": ["python", "fastapi"],
            "itemIds": ["mem-1", "mem-2"],
            "ts": current_time,
        },
        {
            "orgId": "test-org",
            "projectId": "test-project",
            "packId": "pack-2",
            "accepted": False,
            "tags": ["python", "machine_learning"],
            "itemIds": ["mem-3"],
            "ts": current_time,
        },
        {
            "orgId": "test-org",
            "projectId": "test-project",
            "packId": "pack-3",
            "accepted": True,
            "tags": ["fastapi", "web_development"],
            "itemIds": ["mem-4", "mem-5"],
            "ts": current_time,
        },
    ]


@pytest.fixture
def sample_tag_edges():
    """Sample tag edges for testing."""
    current_time = int(time.time())
    return {
        "python": [
            {"to": "fastapi", "w": 0.8, "ts": current_time},
            {"to": "machine_learning", "w": 0.6, "ts": current_time},
        ],
        "fastapi": [
            {"to": "python", "w": 0.8, "ts": current_time},
            {"to": "web_development", "w": 0.4, "ts": current_time},
        ],
        "machine_learning": [
            {"to": "python", "w": 0.6, "ts": current_time},
            {"to": "data_science", "w": 0.7, "ts": current_time},
        ],
    }


@pytest.fixture
def edge_learning_job(mongo_uri):
    """Create edge learning job with test configuration."""
    # Set environment variables for testing
    os.environ["MONGODB_URI"] = mongo_uri
    os.environ["MME_LEARN_ETA"] = "0.1"
    os.environ["MME_LEARN_R"] = "0.05"
    os.environ["MME_LEARN_D"] = "0.02"
    os.environ["MME_LEARN_WMAX"] = "1.0"
    os.environ["MME_LEARN_WINDOW_HOURS"] = "24"
    os.environ["MME_MAX_EDGES_PER_TAG"] = "3"
    
    from app.jobs.edge_learning import EdgeLearningJob
    return EdgeLearningJob()
