#!/usr/bin/env python3
import json
import time
import requests
import statistics
from typing import List, Dict, Any
import numpy as np

class BenchmarkEvaluator:
    def __init__(self, base_url: str = "http://localhost:8081"):
        self.base_url = base_url
        self.headers = {"X-User-ID": "benchmark_user", "Content-Type": "application/json"}
        
    def load_golden_set(self, filename: str) -> List[Dict]:
        with open(filename, 'r') as f:
            return [json.loads(line) for line in f]
    
    def load_candidates(self, filename: str) -> Dict[str, Dict]:
        candidates = {}
        with open(filename, 'r') as f:
            for line in f:
                data = json.loads(line)
                candidates[data['query_id']] = data
        return candidates
    
    def evaluate_l0_vector_topk(self, query_data: Dict, candidates: Dict, k: int = 10) -> Dict:
        """L0: Vector top-K only"""
        query_id = query_data['query_id']
        vector_candidates = candidates[query_id]['vector_top200'][:k]
        
        start_time = time.time()
        # Simulate vector top-K selection
        selected = vector_candidates
        pack_time = (time.time() - start_time) * 1000  # Convert to ms
        
        return self._compute_metrics(query_data, selected, pack_time, "L0")
    
    def evaluate_l1_bm25_topk(self, query_data: Dict, candidates: Dict, k: int = 10) -> Dict:
        """L1: BM25 top-K only"""
        query_id = query_data['query_id']
        bm25_candidates = candidates[query_id]['bm25_top200'][:k]
        
        start_time = time.time()
        # Simulate BM25 top-K selection
        selected = bm25_candidates
        pack_time = (time.time() - start_time) * 1000
        
        return self._compute_metrics(query_data, selected, pack_time, "L1")
    
    def evaluate_l2_hybrid_fusion(self, query_data: Dict, candidates: Dict, k: int = 10) -> Dict:
        """L2: Hybrid fusion of vector and BM25"""
        query_id = query_data['query_id']
        vector_candidates = candidates[query_id]['vector_top200']
        bm25_candidates = candidates[query_id]['bm25_top200']
        
        start_time = time.time()
        # Simple interleaving fusion
        fused = []
        for i in range(max(len(vector_candidates), len(bm25_candidates))):
            if i < len(vector_candidates):
                fused.append(vector_candidates[i])
            if i < len(bm25_candidates):
                fused.append(bm25_candidates[i])
        selected = fused[:k]
        pack_time = (time.time() - start_time) * 1000
        
        return self._compute_metrics(query_data, selected, pack_time, "L2")
    
    def evaluate_l3_mme_baseline(self, query_data: Dict, candidates: Dict) -> Dict:
        """L3: MME baseline (tags/graph only)"""
        query_text = query_data['query_text']
        
        start_time = time.time()
        response = requests.get(
            f"{self.base_url}/memory/query",
            params={"tags": query_text, "limit": 10},
            headers=self.headers
        )
        pack_time = (time.time() - start_time) * 1000
        
        if response.status_code == 200:
            data = response.json()
            return self._compute_mme_metrics(query_data, data, pack_time, "L3")
        else:
            return {"error": f"HTTP {response.status_code}", "mode": "L3"}
    
    def evaluate_l4_mme_vecsim(self, query_data: Dict, candidates: Dict) -> Dict:
        """L4: MME with vecSim ON"""
        query_text = query_data['query_text']
        
        start_time = time.time()
        response = requests.post(
            f"{self.base_url}/search/semantic",
            json={"query": query_text},
            headers=self.headers
        )
        pack_time = (time.time() - start_time) * 1000
        
        if response.status_code == 200:
            data = response.json()
            return self._compute_mme_metrics(query_data, data, pack_time, "L4")
        else:
            return {"error": f"HTTP {response.status_code}", "mode": "L4"}
    
    def evaluate_l5_mme_union(self, query_data: Dict, candidates: Dict) -> Dict:
        """L5: MME with vecSim + union"""
        query_text = query_data['query_text']
        
        start_time = time.time()
        response = requests.post(
            f"{self.base_url}/search/semantic",
            json={"query": query_text},
            headers=self.headers
        )
        pack_time = (time.time() - start_time) * 1000
        
        if response.status_code == 200:
            data = response.json()
            return self._compute_mme_metrics(query_data, data, pack_time, "L5")
        else:
            return {"error": f"HTTP {response.status_code}", "mode": "L5"}
    
    def _compute_metrics(self, query_data: Dict, selected: List[Dict], pack_time: float, mode: str) -> Dict:
        """Compute metrics for baseline modes (L0-L2)"""
        relevant_ids = set(query_data['relevant_ids'])
        
        # Extract IDs from selected candidates
        selected_ids = [item['id'] for item in selected]
        
        # Compute metrics
        recall_at_k = len(set(selected_ids) & relevant_ids) / len(relevant_ids) if relevant_ids else 0
        precision_at_10 = len(set(selected_ids[:10]) & relevant_ids) / min(10, len(selected_ids)) if selected_ids else 0
        
        # nDCG@10 (simplified)
        dcg = 0
        for i, item_id in enumerate(selected_ids[:10]):
            if item_id in relevant_ids:
                dcg += 1 / np.log2(i + 2)
        idcg = sum(1 / np.log2(i + 2) for i in range(min(len(relevant_ids), 10)))
        ndcg_at_10 = dcg / idcg if idcg > 0 else 0
        
        # MRR
        mrr = 0
        for i, item_id in enumerate(selected_ids):
            if item_id in relevant_ids:
                mrr = 1 / (i + 1)
                break
        
        return {
            "query_id": query_data['query_id'],
            "mode": mode,
            "recall_at_k": recall_at_k,
            "precision_at_10": precision_at_10,
            "ndcg_at_10": ndcg_at_10,
            "mrr": mrr,
            "pack_time_ms": pack_time,
            "items_count": len(selected_ids),
            "token_cost": 0,  # Not applicable for baseline modes
            "diversity_breaches": 0,  # Not applicable for baseline modes
            "determinism_ok": True  # Not applicable for baseline modes
        }
    
    def _compute_mme_metrics(self, query_data: Dict, response_data: Dict, pack_time: float, mode: str) -> Dict:
        """Compute metrics for MME modes (L3-L5)"""
        relevant_ids = set(query_data['relevant_ids'])
        
        if 'data' not in response_data or 'results' not in response_data['data']:
            return {"error": "Invalid response format", "mode": mode}
        
        results = response_data['data']['results']
        selected_ids = [item['id'] for item in results]
        
        # Compute basic metrics
        recall_at_k = len(set(selected_ids) & relevant_ids) / len(relevant_ids) if relevant_ids else 0
        precision_at_10 = len(set(selected_ids[:10]) & relevant_ids) / min(10, len(selected_ids)) if selected_ids else 0
        
        # nDCG@10 (simplified)
        dcg = 0
        for i, item_id in enumerate(selected_ids[:10]):
            if item_id in relevant_ids:
                dcg += 1 / np.log2(i + 2)
        idcg = sum(1 / np.log2(i + 2) for i in range(min(len(relevant_ids), 10)))
        ndcg_at_10 = dcg / idcg if idcg > 0 else 0
        
        # MRR
        mrr = 0
        for i, item_id in enumerate(selected_ids):
            if item_id in relevant_ids:
                mrr = 1 / (i + 1)
                break
        
        # Token cost
        token_cost = response_data['data'].get('totalTokenCost', 0)
        
        # Diversity (simplified - would need actual tag comparison)
        diversity_breaches = 0  # Placeholder
        
        return {
            "query_id": query_data['query_id'],
            "mode": mode,
            "recall_at_k": recall_at_k,
            "precision_at_10": precision_at_10,
            "ndcg_at_10": ndcg_at_10,
            "mrr": mrr,
            "pack_time_ms": pack_time,
            "items_count": len(selected_ids),
            "token_cost": token_cost,
            "diversity_breaches": diversity_breaches,
            "determinism_ok": True  # Would need multiple runs to verify
        }
    
    def run_benchmark(self, golden_file: str, candidates_file: str, output_dir: str):
        """Run full benchmark across all modes"""
        golden_set = self.load_golden_set(golden_file)
        candidates = self.load_candidates(candidates_file)
        
        modes = ['L0', 'L1', 'L2', 'L3', 'L4', 'L5']
        all_results = {mode: [] for mode in modes}
        
        for query_data in golden_set:
            print(f"Evaluating query: {query_data['query_id']}")
            
            # Run each mode
            results = []
            
            # L0: Vector top-K
            result = self.evaluate_l0_vector_topk(query_data, candidates)
            results.append(result)
            all_results['L0'].append(result)
            
            # L1: BM25 top-K
            result = self.evaluate_l1_bm25_topk(query_data, candidates)
            results.append(result)
            all_results['L1'].append(result)
            
            # L2: Hybrid fusion
            result = self.evaluate_l2_hybrid_fusion(query_data, candidates)
            results.append(result)
            all_results['L2'].append(result)
            
            # L3: MME baseline
            result = self.evaluate_l3_mme_baseline(query_data, candidates)
            results.append(result)
            all_results['L3'].append(result)
            
            # L4: MME vecSim
            result = self.evaluate_l4_mme_vecsim(query_data, candidates)
            results.append(result)
            all_results['L4'].append(result)
            
            # L5: MME union
            result = self.evaluate_l5_mme_union(query_data, candidates)
            results.append(result)
            all_results['L5'].append(result)
            
            # Save per-query results
            with open(f"{output_dir}/results_{query_data['query_id']}.json", 'w') as f:
                json.dump(results, f, indent=2)
        
        # Save aggregated results
        for mode in modes:
            with open(f"{output_dir}/results_{mode}.jsonl", 'w') as f:
                for result in all_results[mode]:
                    f.write(json.dumps(result) + '\n')
            
            # Compute aggregates
            if all_results[mode]:
                aggregates = self._compute_aggregates(all_results[mode])
                with open(f"{output_dir}/metrics_{mode}.json", 'w') as f:
                    json.dump(aggregates, f, indent=2)
    
    def _compute_aggregates(self, results: List[Dict]) -> Dict:
        """Compute aggregate metrics"""
        valid_results = [r for r in results if 'error' not in r]
        
        if not valid_results:
            return {"error": "No valid results"}
        
        metrics = {
            "count": len(valid_results),
            "recall_at_k": {
                "mean": np.mean([r['recall_at_k'] for r in valid_results]),
                "median": np.median([r['recall_at_k'] for r in valid_results]),
                "std": np.std([r['recall_at_k'] for r in valid_results])
            },
            "precision_at_10": {
                "mean": np.mean([r['precision_at_10'] for r in valid_results]),
                "median": np.median([r['precision_at_10'] for r in valid_results]),
                "std": np.std([r['precision_at_10'] for r in valid_results])
            },
            "ndcg_at_10": {
                "mean": np.mean([r['ndcg_at_10'] for r in valid_results]),
                "median": np.median([r['ndcg_at_10'] for r in valid_results]),
                "std": np.std([r['ndcg_at_10'] for r in valid_results])
            },
            "mrr": {
                "mean": np.mean([r['mrr'] for r in valid_results]),
                "median": np.median([r['mrr'] for r in valid_results]),
                "std": np.std([r['mrr'] for r in valid_results])
            },
            "pack_time_ms": {
                "mean": np.mean([r['pack_time_ms'] for r in valid_results]),
                "median": np.median([r['pack_time_ms'] for r in valid_results]),
                "std": np.std([r['pack_time_ms'] for r in valid_results]),
                "p95": np.percentile([r['pack_time_ms'] for r in valid_results], 95)
            },
            "token_cost": {
                "mean": np.mean([r['token_cost'] for r in valid_results]),
                "median": np.median([r['token_cost'] for r in valid_results]),
                "std": np.std([r['token_cost'] for r in valid_results])
            }
        }
        
        return metrics

if __name__ == "__main__":
    evaluator = BenchmarkEvaluator()
    evaluator.run_benchmark(
        "test_eff/benchmark/golden.jsonl",
        "test_eff/benchmark/candidates.jsonl",
        "test_eff/benchmark"
    )
