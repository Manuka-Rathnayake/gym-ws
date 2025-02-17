#include <ESP8266WiFi.h>
#include <WebSocketsClient.h>
#include <ArduinoJson.h>

#define RELAY D1
#define BUZZER D8   
#define SWITCH D2   
#define RED_LED D5
#define GREEN_LED D6
#define BLUE_LED D7

const char* ssid = "DataStream_2.4";       
const char* password = "pwddd";  

const char* websocket_host = "192.168.1.6";  
const int websocket_port = 8080;              
const char* websocket_path = "/ws";

WebSocketsClient webSocket;
String gymID = "G124";  
bool doorUnlocking = false;

void webSocketEvent(WStype_t type, uint8_t *payload, size_t length) {
    switch (type) {
        case WStype_DISCONNECTED:
            Serial.println("⚠️ WebSocket Disconnected! Attempting Reconnect...");
            break;

        case WStype_CONNECTED: {
            Serial.println("✅ WebSocket Connected to Local Server!");

            DynamicJsonDocument doc(256);
            String jsonStr;

            doc["type"] = "REGISTER";
            doc["gymID"] = gymID;

            serializeJson(doc, jsonStr);
            Serial.print("📡 Sending JSON Gym ID: ");
            Serial.println(jsonStr);
            webSocket.sendTXT(jsonStr);
            break;
        }

        case WStype_TEXT:
            Serial.print("📩 Received from server: ");
            Serial.println((char*)payload);

            if (String((char*)payload) == "UNLOCK") {
                Serial.println("🚪 WebSocket Unlock Command Received - Unlocking Door...");
                
                
                delay(50);  

                unlockDoor();
            }
            break;

        case WStype_ERROR:
            Serial.println("❌ WebSocket Error!");
            break;

        default:
            Serial.println("ℹ️ WebSocket Event Received.");
            break;
    }
}

void setup() {
    Serial.begin(115200);
    Serial.println("\n🔹 ESP8266 Booting...");

    pinMode(RELAY, OUTPUT);
    pinMode(BUZZER, OUTPUT);
    pinMode(SWITCH, INPUT_PULLUP);  
    pinMode(RED_LED, OUTPUT);
    pinMode(GREEN_LED, OUTPUT);
    pinMode(BLUE_LED, OUTPUT);

    digitalWrite(RELAY, HIGH);
    digitalWrite(RED_LED, HIGH);
    digitalWrite(GREEN_LED, LOW);
    digitalWrite(BUZZER, LOW);

    // 🔹 Connect to WiFi
    Serial.printf("📡 Connecting to WiFi: %s\n", ssid);
    WiFi.begin(ssid, password);
    int attempt = 0;

    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
        attempt++;
        if (attempt > 30) {  
            Serial.println("\n❌ WiFi Connection Failed! Restarting ESP...");
            ESP.restart();
        }
    }

    Serial.println("\n✅ WiFi Connected!");
    Serial.printf("🌐 IP Address: %s\n", WiFi.localIP().toString().c_str());

    // 🔹 Connect to Local WebSocket Server
    Serial.println("🔄 Connecting to Local WebSocket Server...");
    webSocket.begin(websocket_host, websocket_port, websocket_path);
    webSocket.onEvent(webSocketEvent);
    webSocket.setReconnectInterval(5000);
}

void loop() {
    webSocket.loop();

    static unsigned long lastPing = 0;
    if (millis() - lastPing > 30000) {  
        lastPing = millis();
        webSocket.sendTXT("PING");
        Serial.println("📡 Sent WebSocket Keepalive Ping");
    }

    // Prevent multiple IR triggers
    if (!doorUnlocking && digitalRead(SWITCH) == LOW) {
        Serial.println("🚪 IR Sensor Detected - Unlocking Door...");
        doorUnlocking = true;
        unlockDoor();
        delay(5000);  // Prevent multiple triggers
        doorUnlocking = false;
    }
}

void unlockDoor() {
    Serial.println("🔓 Unlocking Door...");
    digitalWrite(RELAY, LOW);
    digitalWrite(GREEN_LED, HIGH);
    digitalWrite(RED_LED, LOW);

    playUnlockBeep();
    
    delay(3000);

    Serial.println("🔒 Locking Door...");
    digitalWrite(RELAY, HIGH);
    digitalWrite(GREEN_LED, LOW);
    digitalWrite(RED_LED, HIGH);
}

void playUnlockBeep() {
    Serial.println("🔊 Playing unlock beep...");
    
    for (int i = 0; i < 3; i++) {
        digitalWrite(BUZZER, HIGH);
        delay(100);
        digitalWrite(BUZZER, LOW);
        delay(100);
    }
}
