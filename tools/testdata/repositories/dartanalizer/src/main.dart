// Unused import
import 'dart:io';

// Function with unused parameter and missing return type
function1(String unusedParam) {
  var x = 42;
  return;
}

// Class with fields that should be final
class BadClass {
  String mutableField = "test";
  var undefinedType = 123;
  
  // Method with unused parameter
  void doSomething(int unused) {
    print("doing nothing");
  } 
}

void main() {
  // Unused variable
  var unused = 100;
  
  // Variable without type annotation
  var something = "hello";
  
  // Dead code after return
  if (true) {
    return;
    print("unreachable");
  }
  
  // Using dynamic when a specific type would work
  dynamic number = 42;
  number = "not a number anymore";
}
