public class Test {
    private String unusedField; // PMD: UnusedPrivateField

    public void testMethod() {
        if (true) { // PMD: UnconditionalIfStatement
            System.out.println("This is always true");
        }
        
        String str = null;
        if (str.equals("test")) { // PMD: NullPointerException
            System.out.println("This will cause NPE");
        }
    }
} 