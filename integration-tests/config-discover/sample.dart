/// A simple Dart file to test config discovery functionality.

void main() {
  helloWorld();
  final result = addNumbers(5, 3);
  print('5 + 3 = $result');
}

void helloWorld() {
  print('Hello, World!');
}

int addNumbers(int a, int b) {
  return a + b;
} 