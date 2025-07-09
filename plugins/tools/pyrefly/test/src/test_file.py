#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test file for pyrefly analysis
"""

def add_numbers(a: int, b: int) -> int:
    return a + b

def call_with_wrong_type():
    # This should trigger a type error in pyrefly
    return add_numbers("one", 2)

def missing_return() -> int:
    # This function is missing a return statement
    pass

def wrong_return_type() -> str:
    # This function returns an int instead of a str
    return 123

if __name__ == "__main__":
    call_with_wrong_type()
    missing_return()
    wrong_return_type() 