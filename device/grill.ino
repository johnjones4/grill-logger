#include <WiFi.h>
#include <HTTPClient.h>
#include "secrets.h"

// the value of the 'other' resistor
#define SERIESRESISTOR 100000    
 
// What pin to connect the sensor to
#define THERMISTORPIN1 A0 
#define THERMISTORPIN2 A1

// resistance at 25 degrees C
#define THERMISTORNOMINAL 100000      
// temp. for nominal resistance (almost always 25 C)
#define TEMPERATURENOMINAL 25  
// The beta coefficient of the thermistor (usually 3000-4000)
#define BCOEFFICIENT 3950
// Samples to take per reading
#define NUMSAMPLES 10

#define LED_WORKING 14
#define LED_READING 32
#define LED_TRANSMIT 15
#define LED_ERROR 33
#define N_LEDS 4

int leds[] = {LED_WORKING, LED_READING, LED_TRANSMIT, LED_ERROR};

void setup(void) {
  Serial.begin(115200);
  for (int i = 0; i < N_LEDS; i++) {
    pinMode(leds[i], OUTPUT);
  }
  setStatusLed(LED_WORKING, true);
}

float takeReading(int pin) {
  setStatusLed(LED_READING, true);
  int samples[NUMSAMPLES];
  int total;
  for (int i=0; i<NUMSAMPLES; i++) {
    total += analogRead(pin);
    delay(10);
  }
  setStatusLed(LED_READING, false);
  return ((float)total/(float)NUMSAMPLES);
}

float calculateTempCelsius(float reading) {
  reading = (4095 / reading)  - 1;     // (1023/ADC - 1) 
  reading = SERIESRESISTOR / reading;  // 100K / (1023/ADC - 1)
  float steinhart;
  steinhart = reading / THERMISTORNOMINAL;     // (R/Ro)
  steinhart = log(steinhart);                  // ln(R/Ro)
  steinhart /= BCOEFFICIENT;                   // 1/B * ln(R/Ro)
  steinhart += 1.0 / (TEMPERATURENOMINAL + 273.15); // + (1/To)
  steinhart = 1.0 / steinhart;                 // Invert
  steinhart -= 273.15; 
  return steinhart;
}

bool connectWifi() {
//  WiFi.config(STATIC_IP, STATIC_GATEWAY, STATIC_SUBNET, STATIC_DNS); 
  
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  const unsigned long start = millis();
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.println("Establishing connection to WiFi..");
    if (millis() - start > 60000) {
      Serial.println("WiFi timeout ...");
      setStatusLed(LED_ERROR, true);
      return false;
    }
  }
 
  Serial.println("Connected to network");

  return true;
}


void logTemperatures(float celsuisTemp1, float celsuisTemp2) {
  setStatusLed(LED_TRANSMIT, true);
  if (WiFi.status() != WL_CONNECTED) {
    if (!connectWifi()) {
      return;
    }
  }
  int now = millis();
  HTTPClient http;
  http.begin(POST_URL);
  http.addHeader("Content-Type", "application/json");
  char postData[512];
  snprintf(postData, sizeof(postData), "{\"uptime\":%d,\"meatTemp\":%f,\"smokeTemp\":%f}", now, celsuisTemp1, celsuisTemp2);
  int httpResponseCode = http.POST(postData);
  Serial.println(postData);
  Serial.println(httpResponseCode);
  if (httpResponseCode != 200) {
    Serial.printf("Bad HTTP response code %d\n", httpResponseCode);
    setStatusLed(LED_ERROR, true);
    return;
  }
  setStatusLed(LED_TRANSMIT, false);
  setStatusLed(LED_ERROR, false);
}

void setStatusLed(int led, bool ledOn) {
  if (ledOn) {
    digitalWrite(led, HIGH);
  } else {
    digitalWrite(led, LOW);
  }
}
 
void loop(void) {
  int startTime = millis();
  
  float reading1 = takeReading(THERMISTORPIN1);
  float reading2 = takeReading(THERMISTORPIN2);
  float temp1 = calculateTempCelsius(reading1);
  float temp2 = calculateTempCelsius(reading2);

  logTemperatures(temp1, temp2);

  int endTime = millis();
  
  delay(20000 - (endTime - startTime));
}
