// Missing semicolon
const x = 10
const y = 20

// Unused variable
const unused = 'test'

// Console statement
console.log('test')

// Unused parameter
function testFunction(param) {
    return true
}

// Missing 'use strict'
function strictFunction() {
    undeclared = 'test'
}

// Inconsistent spacing
if(x>0){
    console.log('positive')
}

// Unreachable code
function unreachable() {
    return true
    console.log('never reached')
}

// Missing return type
function noReturnType() {
    return 'test'
}

// Using var instead of const/let
var oldStyle = 'test'

// Using == instead of ===
if (x == '10') {
    console.log('loose equality')
}

// Using eval
eval('console.log("dangerous")')

// Using with statement
with (Math) {
    console.log(PI)
}

// Using arguments object
function useArguments() {
    console.log(arguments)
}

// Using this in arrow function
const arrowWithThis = () => {
    console.log(this)
}

// Using prototype
function PrototypeTest() {}
PrototypeTest.prototype.test = function() {
    console.log('prototype method')
}

// Using new without assignment
new Date()

// Using void operator
void 0

// Using debugger statement
debugger

// Using label
label: {
    console.log('labeled statement')
}

// Using octal literal
const octal = 0o123 