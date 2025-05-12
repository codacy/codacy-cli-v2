// Unused import
import 'dart:math';

// Unused variable
var unusedVar = 42;

// Function with missing return type and parameter type
foo(bar) {
  print(bar);
}

// Function with always true condition
void alwaysTrue() {
  if (1 == 1) {
    print('This is always true');
  }
}

// Function with a deprecated member usage
@deprecated
void oldFunction() {
  print('This function is deprecated');
}

void main() {
  foo('test');
  alwaysTrue();
  oldFunction();
} 