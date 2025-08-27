import json
import requests
import time

BASE_URL = "http://localhost:8081"
HEADERS = {
    "Content-Type": "application/json",
    "X-User-ID": "stress-test",
    "X-Org-ID": "stress-test"
}

def test_edge_case(name, content, expected_behavior="success"):
    print(f"\n🧪 TEST: {name}")
    print(f"Content length: {len(content)} chars, {len(content.split())} words")
    
    payload = {"prompt": content}
    
    try:
        start_time = time.time()
        response = requests.post(f"{BASE_URL}/tags/extract", json=payload, headers=HEADERS, timeout=10)
        end_time = time.time()
        
        if response.status_code == 200:
            data = response.json()
            tags = data.get('tags', [])
            tag_count = len(tags) if tags else 0
            print(f"✅ SUCCESS: {tag_count} tags in {(end_time-start_time)*1000:.1f}ms")
            if tags:
                print(f"Sample tags: {', '.join(tags[:10])}")
            return True, tags
        else:
            error_data = response.json()
            print(f"❌ FAILED: {error_data.get('error', 'Unknown error')}")
            return False, error_data.get('error')
            
    except requests.exceptions.Timeout:
        print("⏰ TIMEOUT: Request took longer than 10 seconds")
        return False, "timeout"
    except Exception as e:
        print(f"💥 EXCEPTION: {str(e)}")
        return False, str(e)

# Test cases
if __name__ == "__main__":
    print("Starting comprehensive breaking point tests...")
