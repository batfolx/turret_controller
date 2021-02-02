  #include <ArduinoJson.h>
#include <Servo.h>

// angles
uint8_t HIGH_ANGLE = 165;
int8_t LOW_ANGLE = 15;
uint8_t MID_ANGLE = 90;

uint8_t RETRACTED = 180;
uint8_t FORWARD = 0;

// use PWM pins for controlling servos of movement
uint8_t base_servo_pin = 9;
uint8_t top_pin = 10; 
uint8_t trigger_pin = 11;

// pins for controlling the shooting mechanism
int8_t ENA = A0;
int8_t EN1 = A1;
int8_t EN2 = 7;

int8_t ENB = A3;
int8_t EN3 = A2;
int8_t EN4 = 8;




bool forwards = false;
bool arm = true;

unsigned long last_time_scanned = millis() / 1000;


Servo base_servo;
Servo top_servo;
Servo trigger_servo;

// json document to store incoming JSON data from the raspberry pi 
// 512 bytes should be good enough
DynamicJsonDocument doc(512);



void setup() {
  Serial.begin(9600);

  initialize_pins();
  center_servos();
  
}


/**
 * Centers the servos 
 */
void center_servos() {

  // center the camera and base servo
  top_servo.write(MID_ANGLE);
  base_servo.write(MID_ANGLE);

  trigger_servo.write(RETRACTED);
  delay(1000);
}

void initialize_pins() {
  

  // set up the motor
  pinMode(ENA, OUTPUT);
  pinMode(EN1, OUTPUT);
  pinMode(EN2, OUTPUT);

  pinMode(ENB, OUTPUT);
  pinMode(EN3, OUTPUT);
  pinMode(EN4, OUTPUT);

  // attach all of the servos, we only need 3 
  base_servo.attach(base_servo_pin);
  top_servo.attach(top_pin);
  trigger_servo.attach(trigger_pin);

}



// spins up ther motors to shoot the darts
void spin_motors() {
 
   digitalWrite(EN1, LOW);
   digitalWrite(EN2, HIGH);


   digitalWrite(EN3, HIGH);
   digitalWrite(EN4, LOW);

   analogWrite(ENA, 500);
   analogWrite(ENB, 500);
}

void fire() {
  spin_motors();
  trigger_servo.write(FORWARD);
  delay(500);
  trigger_servo.write(RETRACTED);
  delay(500);
}

void stop_motors() {


  analogWrite(ENA, 0);
  analogWrite(ENB, 0);
  
}

void loop() {
  read_serial();
}



void read_serial() {

  if (Serial.available() > 0) {
      Serial.println("Found some stuff");
      String json = Serial.readStringUntil('\0');
      DeserializationError err = deserializeJson(doc, json.c_str());
      
      // just continue to the next loop. fugget about it.
      if (err) return;
      last_time_scanned = millis() / 1000;
      Serial.print("This is json from Serial: ");
      Serial.println(json);

      
      int8_t side_value = doc["side"];
      int8_t base_value = doc["base"];
      bool fire_gun = doc["fire"];

      
      // turn both servos at the same time
      turn_servo(side_value, &top_servo);
      turn_servo(base_value, &base_servo);

      // if main controller says to fire, we will fire
      if (fire_gun) fire();
      else stop_motors();

      delay(25);
      
  } else {

      
      // if 5 seconds has passed since last log time, start scanning
      if ( ((millis() / 1000) - last_time_scanned) > 5 ) scan(&base_servo); 
      
    
  }
  
}

void scan(Servo *s) {
  center_cam();
  int angle = s->read();
  int8_t delta = forwards ? 1 : -1;
  int change = angle + delta;
  if ((change > HIGH_ANGLE - 15) || (change < LOW_ANGLE + 15)) forwards = !forwards;
  s->write(change);
  delay(25);
  
}

void center_cam() {
  top_servo.write(MID_ANGLE);
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
