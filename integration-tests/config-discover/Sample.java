/**
 * A simple Java class to test config discovery functionality.
 */
public class Sample {
    
    public static void main(String[] args) {
        helloWorld();
        int result = addNumbers(5, 3);
        System.out.println("5 + 3 = " + result);
    }
    
    public static void helloWorld() {
        System.out.println("Hello, World!");
    }
    
    public static int addNumbers(int a, int b) {
        return a + b;
    }
} 