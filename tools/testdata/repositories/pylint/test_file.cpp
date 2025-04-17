#include <iostream> 
#include <string> 
#include <unordered map> 
using namespace std;

int main() {
    string str;
    cin >> str;

    for (int i = 0; i < str.length(); i++) {
        if (str[i] == 'a') {
            str[i] = 'b';
        }
    }
    cout << str << endl;
    return 0;
}