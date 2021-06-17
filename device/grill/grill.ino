#include <LiquidCrystal.h>
#include <WiFi.h>

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
 
LiquidCrystal lcd(14, 32, 15, 33, 27, 12);

void setup(void) {
  Serial.begin(115200);
//  lcd.begin(16, 2);
}

float takeReading(int pin) {
  int samples[NUMSAMPLES];
  int total;
  for (int i=0; i<NUMSAMPLES; i++) {
    total += analogRead(pin);
    delay(10);
  }
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

float celsuisToFahrenheit(float celsius) {
  return (celsius * 9 / 5) + 32;
}

float displayTemperatureLine(int line, float celsuisTemp) {
  float fahrenheit = celsuisToFahrenheit(celsuisTemp);
  char buffer[16];
  memset(buffer, ' ', 16);
  sprintf(buffer, "%.2f *F", fahrenheit);
  lcd.setCursor(line, 0);
  lcd.print(buffer);
}

float displayTemperatures(float celsuisTemp1, float celsuisTemp2) {
  displayTemperatureLine(0, celsuisTemp1);
  displayTemperatureLine(1, celsuisTemp2);
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
      return false;
    }
  }
 
  Serial.println("Connected to network");

  return true;
}


float logTemperatures(float celsuisTemp1, float celsuisTemp2) {
  
}
 
void loop(void) {
  int startTime = millis();
  
  float reading1 = takeReading(THERMISTORPIN1);
  float reading2 = takeReading(THERMISTORPIN2);
  float temp1 = calculateTempCelsius(reading1);
  float temp2 = calculateTempCelsius(reading2);

  displayTemperatures(temp1, temp2);
  logTemperatures(temp1, temp2);

  int endTime = millis();
  
  delay(10000 - (endTime - startTime));
}
