#!/usr/bin/env python3
"""
A simple Python file to test config discovery functionality.
"""

def hello_world():
    """Print hello world message."""
    print("Hello, World!")

def add_numbers(a, b):
    """Add two numbers and return the result."""
    return a + b

if __name__ == "__main__":
    hello_world()
    result = add_numbers(5, 3)
    print(f"5 + 3 = {result}") 