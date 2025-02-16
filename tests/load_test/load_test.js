import { check } from "k6";
import http from "k6/http";

// k6 run load_test.js --summary-export=result.json
export let options = {
    scenarios: {
        constant_rps: {
            executor: 'constant-arrival-rate',
            rate: 250, // currently 4 requests in func --> ~1000req/sec.
            timeUnit: '1s',
            duration: '100s', // be careful, user's money can finish
            preAllocatedVUs: 300,
            maxVUs: 100000,
        },
    },
    thresholds: {
        // Percent of errors - currently need <0.01%.
        "http_req_failed": ["rate<0.0001"],
        // Response duration: median must be 
        // less than 50ms.
        "http_req_duration": ["med<50"],
    }
};

export default function() {
    let username = `testuser${__VU}`;
    // 1. Auth (POST /api/auth)
    let authPayload = JSON.stringify({
        username: username,
        password: "testpassword",
    });
    let authParams = { headers: { "Content-Type": "application/json"} };

    let authRes = http.post("http://localhost:8080/api/auth", authPayload, authParams);
    check(authRes, {
        "auth status is 200": (r) => r.status == 200,
        "token exists": (r) => r.json("token") !== undefined, 
    });

    let token = authRes.json("token");
    let params = {
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + token,
        }
    };

    // 2. User info (GET /api/info)
    let infoRes = http.get("http://localhost:8080/api/info", params);
    check (infoRes, {
        "info status is 200": (r) => r.status === 200,
    });

    // 3. Send Coins to other user (POST /api/sendCoin)
    let sendCoinPayload = JSON.stringify({
        toUser: "testuser",
        amount: 1,
    })
    let sendCoinRes = http.post("http://localhost:8080/api/sendCoin", sendCoinPayload, params);
    check(sendCoinRes, {
        "send status is 200": (r) => r.status === 200, 
    });
    
    // 4. Buy Items (GET /api/buy/:item)
    let item = "pen";
    let buyItemRes = http.get(`http://localhost:8080/api/buy/${item}`, params);
    check(buyItemRes, {
        "buy status is 200": (r) => r.status === 200,
    });
}
