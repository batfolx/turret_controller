#include <ArduinoJson.h>
#include <Servo.h>

uint8_t HIGH_ANGLE = 165;
int8_t LOW_ANGLE = 15;
uint8_t MID_ANGLE = 90;

// use PWM pins for controlling servos
uint8_t base_servo_pin = 9;
uint8_t top_pin = 10;

uint8_t left_pin= 3;
uint8_t right_pin = 5;

bool forwards = false;
bool arm = false;

unsigned long last_time_scanned = millis() / 1000;


Servo base_servo;
Servo top_servo;
Servo right;
Servo left;

// json document to store incoming JSON data from the raspberry pi 
// 512 bytes should be good enough
DynamicJsonDocument doc(512);



void setup() {
  
  Serial.begin(9600);
  base_servo.attach(base_servo_pin);
  top_servo.attach(top_pin);

  left.attach(left_pin);
  right.attach(right_pin);

  
  Serial.println("Lets go!");
  base_servo.write(MID_ANGLE);
  top_servo.write(MID_ANGLE);
  left.write(180);
  right.write(angle_inv(180));
  
  
  
  delay(500);

}

void loop() {
  read_serial();
}

// right servo is inverted so transform
int angle_inv(int angle) {

   return abs(180 - angle);
    
}


void read_serial() {

  if (Serial.available() > 0) {
      Serial.println("Found some stuff");
      String json = Serial.readStringUntil('\0');
      DeserializationError err = deserializeJson(doc, json.c_str());
      Serial.print("This is json from Serial: ");
      Serial.println(json);
      // just continue to the next loop. fugget about it.
      if (err) return;
      last_time_scanned = millis() / 1000;


      
      int16_t side_value = doc["side"];
      int16_t base_value = doc["base"];

      Serial.println(side_value);

      if (arm) {
        calibrate_servo(&top_servo, &left, false);
        calibrate_servo(&top_servo, &right, true);
        delay(25);
        arm = false;
      }

      // turn both servos at the same time
      turn_servo(side_value, &top_servo);
      turn_servo(side_value, &left);
      turn_servo(-side_value, &right);

      turn_servo(base_value, &base_servo);
      delay(25);
      
  } else {
      if (!arm) {
        left.write(180);
        right.write(angle_inv(180));
        delay(50);
        arm = true;
      }
      
      // if 5 seconds has passed since last log time, start scanning
      if ( ((millis() / 1000) - last_time_scanned) > 5 )scan(&base_servo);
    
  }
  
}

void scan(Servo *s) {

  int angle = s->read();
  int8_t delta = forwards ? 1 : -1;
  int change = angle + delta;
  if ((change > HIGH_ANGLE - 15) || (change < LOW_ANGLE + 15)) forwards = !forwards;
  s->write(change);
  delay(25);
  
}

void calibrate_servo(Servo *ref, Servo *to_cali, bool right) {

  if (right)to_cali->write(angle_inv(ref->read()));
  else to_cali->write(ref->read());
    
}

void turn_servo(int16_t value, Servo *s) {

  // gets the current angle of the servo, from 0 - 180
  // meaning we can store in uint8_t (0-255)
  uint8_t current_angle = s->read();

  // get new angle, values from 2^(16-1)
  int16_t new_angle;

  // negative value, subtract from new angle to rotate servo to new angle (left)
  if (value < 0) {

    // check to make sure we don't write under 0. new_angle can hold negative values so check
    // if subtracted value is less than zero
    new_angle = current_angle - abs(value);

    // reset servo if goes below zero
    if (new_angle < LOW_ANGLE) {
     s->write(MID_ANGLE);
     delay(500); 
      
    }
    else s->write(new_angle);
    
  // positive, add to current angle
  } else {

    // check if new angle greater than max angle of 180
    // and only write 180 if that is true
    new_angle = current_angle + value;

    // reset servo if angle goes above 180
    if (new_angle > HIGH_ANGLE) {

       s->write(HIGH_ANGLE);
       delay(500);
      
    }
    else s->write(new_angle);
    
    
  }
  
  
}
