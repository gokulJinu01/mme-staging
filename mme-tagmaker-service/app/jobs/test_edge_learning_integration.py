"""
Integration tests for Edge Learning Job using real MongoDB.

Tests EMA-based edge weight updates and pruning functionality with actual database operations.
"""

import time
import pytest
from app.tests.conftest import seed_pack_events, seed_tag_edges, get_tag_edges, get_pack_events


class TestEdgeLearningIntegration:
    """Integration tests for edge learning with real MongoDB."""

    def test_edge_learning_with_real_data(self, mongo_db, edge_learning_job, sample_pack_events, sample_tag_edges):
        """Test edge learning with real pack events and tag edges."""
        org_id = "test-org"
        
        # Seed initial tag edges
        seed_tag_edges(mongo_db, org_id, sample_tag_edges)
        
        # Verify initial state
        initial_edges = get_tag_edges(mongo_db, org_id, "python")
        assert len(initial_edges) == 2
        assert initial_edges[0]["to"] == "fastapi"
        assert initial_edges[0]["w"] == 0.8
        
        # Seed pack events
        seed_pack_events(mongo_db, org_id, sample_pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        
        # Verify results
        assert result["updated"] > 0
        assert result["duration_seconds"] > 0
        
        # Verify edge updates
        updated_edges = get_tag_edges(mongo_db, org_id, "python")
        assert len(updated_edges) <= 3  # Should respect max edges per tag
        
        # Check that weights have been updated (should be different from initial)
        fastapi_edge = next((edge for edge in updated_edges if edge["to"] == "fastapi"), None)
        assert fastapi_edge is not None
        # Weight should have changed due to EMA updates
        assert fastapi_edge["w"] != 0.8

    def test_ema_math_with_acceptance(self, mongo_db, edge_learning_job):
        """Test EMA math for accepted pack events."""
        org_id = "test-org"
        
        # Create simple test data
        initial_edges = {
            "python": [
                {"to": "fastapi", "w": 0.5, "ts": int(time.time())},
            ]
        }
        seed_tag_edges(mongo_db, org_id, initial_edges)
        
        # Create accepted pack event
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": True,  # Acceptance should increase weight
                "tags": ["python", "fastapi"],
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        assert result["updated"] > 0
        
        # Verify weight increased due to acceptance
        updated_edges = get_tag_edges(mongo_db, org_id, "python")
        fastapi_edge = next((edge for edge in updated_edges if edge["to"] == "fastapi"), None)
        assert fastapi_edge is not None
        assert fastapi_edge["w"] > 0.5  # Should increase due to acceptance

    def test_ema_math_with_rejection(self, mongo_db, edge_learning_job):
        """Test EMA math for rejected pack events."""
        org_id = "test-org"
        
        # Create simple test data
        initial_edges = {
            "python": [
                {"to": "fastapi", "w": 0.5, "ts": int(time.time())},
            ]
        }
        seed_tag_edges(mongo_db, org_id, initial_edges)
        
        # Create rejected pack event
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": False,  # Rejection should decrease weight
                "tags": ["python", "fastapi"],
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        assert result["updated"] > 0
        
        # Verify weight decreased due to rejection
        updated_edges = get_tag_edges(mongo_db, org_id, "python")
        fastapi_edge = next((edge for edge in updated_edges if edge["to"] == "fastapi"), None)
        assert fastapi_edge is not None
        assert fastapi_edge["w"] < 0.5  # Should decrease due to rejection

    def test_weight_clipping(self, mongo_db, edge_learning_job):
        """Test that weights are clipped to [0, w_max] range."""
        org_id = "test-org"
        
        # Create edge with very high weight
        initial_edges = {
            "python": [
                {"to": "fastapi", "w": 2.0, "ts": int(time.time())},  # Above w_max
            ]
        }
        seed_tag_edges(mongo_db, org_id, initial_edges)
        
        # Create pack event
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": True,
                "tags": ["python", "fastapi"],
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        assert result["updated"] > 0
        
        # Verify weight is clipped
        updated_edges = get_tag_edges(mongo_db, org_id, "python")
        fastapi_edge = next((edge for edge in updated_edges if edge["to"] == "fastapi"), None)
        assert fastapi_edge is not None
        assert fastapi_edge["w"] <= 1.0  # Should be clipped to w_max
        assert fastapi_edge["w"] >= 0.0  # Should not be negative

    def test_top_m_pruning(self, mongo_db, edge_learning_job):
        """Test that edges are pruned to top-M per tag."""
        org_id = "test-org"
        
        # Create more edges than max_edges_per_tag (3)
        initial_edges = {
            "python": [
                {"to": "fastapi", "w": 0.8, "ts": int(time.time())},
                {"to": "machine_learning", "w": 0.6, "ts": int(time.time())},
                {"to": "data_science", "w": 0.4, "ts": int(time.time())},
                {"to": "web_development", "w": 0.2, "ts": int(time.time())},  # Should be pruned
            ]
        }
        seed_tag_edges(mongo_db, org_id, initial_edges)
        
        # Create pack events to trigger updates
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": True,
                "tags": ["python", "fastapi"],
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        assert result["updated"] > 0
        
        # Verify pruning to top-M
        updated_edges = get_tag_edges(mongo_db, org_id, "python")
        assert len(updated_edges) <= 3  # Should respect max_edges_per_tag
        
        # Verify edges are sorted by weight (descending)
        weights = [edge["w"] for edge in updated_edges]
        assert weights == sorted(weights, reverse=True)

    def test_bidirectional_edges(self, mongo_db, edge_learning_job):
        """Test that edges are updated bidirectionally."""
        org_id = "test-org"
        
        # Create initial edges
        initial_edges = {
            "python": [
                {"to": "fastapi", "w": 0.5, "ts": int(time.time())},
            ],
            "fastapi": [
                {"to": "python", "w": 0.5, "ts": int(time.time())},
            ]
        }
        seed_tag_edges(mongo_db, org_id, initial_edges)
        
        # Create pack event
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": True,
                "tags": ["python", "fastapi"],
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        assert result["updated"] > 0
        
        # Verify both directions are updated
        python_edges = get_tag_edges(mongo_db, org_id, "python")
        fastapi_edges = get_tag_edges(mongo_db, org_id, "fastapi")
        
        python_to_fastapi = next((edge for edge in python_edges if edge["to"] == "fastapi"), None)
        fastapi_to_python = next((edge for edge in fastapi_edges if edge["to"] == "python"), None)
        
        assert python_to_fastapi is not None
        assert fastapi_to_python is not None
        assert python_to_fastapi["w"] > 0.5  # Should be updated
        assert fastapi_to_python["w"] > 0.5  # Should be updated

    def test_edge_learning_with_no_events(self, mongo_db, edge_learning_job, sample_tag_edges):
        """Test edge learning when no pack events exist."""
        org_id = "test-org"
        
        # Seed tag edges but no pack events
        seed_tag_edges(mongo_db, org_id, sample_tag_edges)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        
        # Should complete without errors but no updates
        assert result["updated"] == 0
        assert result["duration_seconds"] > 0

    def test_edge_learning_with_insufficient_tags(self, mongo_db, edge_learning_job):
        """Test edge learning with pack events that have insufficient tags."""
        org_id = "test-org"
        
        # Create pack event with only one tag (insufficient for pairs)
        pack_events = [
            {
                "orgId": org_id,
                "projectId": "test-project",
                "packId": "pack-1",
                "accepted": True,
                "tags": ["python"],  # Only one tag
                "itemIds": ["mem-1"],
                "ts": int(time.time()),
            }
        ]
        seed_pack_events(mongo_db, org_id, pack_events)
        
        # Run edge learning
        result = edge_learning_job.run(org_id=org_id, hours=24)
        
        # Should complete without errors but no updates (no tag pairs)
        assert result["updated"] == 0
        assert result["duration_seconds"] > 0


class TestEdgeLearningReplayEndpoint:
    """Integration tests for edge learning replay endpoint."""

    def test_replay_endpoint_with_real_data(self, mongo_db, sample_pack_events, sample_tag_edges):
        """Test the replay endpoint with real data."""
        org_id = "test-org"
        
        # Seed data
        seed_tag_edges(mongo_db, org_id, sample_tag_edges)
        seed_pack_events(mongo_db, org_id, sample_pack_events)
        
        # Import the replay function
        from app.jobs.edge_learning import run_edge_learning_for_org
        
        # Run replay
        result = run_edge_learning_for_org(org_id, hours=24)
        
        # Verify response structure
        assert "updated" in result
        assert "pruned" in result
        assert "duration_seconds" in result
        assert result["updated"] > 0
        assert result["duration_seconds"] > 0

    def test_replay_endpoint_with_custom_hours(self, mongo_db, sample_pack_events, sample_tag_edges):
        """Test the replay endpoint with custom hours parameter."""
        org_id = "test-org"
        
        # Seed data
        seed_tag_edges(mongo_db, org_id, sample_tag_edges)
        seed_pack_events(mongo_db, org_id, sample_pack_events)
        
        from app.jobs.edge_learning import run_edge_learning_for_org
        
        # Run replay with custom hours
        result = run_edge_learning_for_org(org_id, hours=12)
        
        # Verify response
        assert "updated" in result
        assert "pruned" in result
        assert "duration_seconds" in result
        assert result["duration_seconds"] > 0
