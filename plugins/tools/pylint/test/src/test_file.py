#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Test file for pylint analysis
"""

import os
import sys

def unused_variable():
    """Function with unused variable"""
    x = 10  # pylint: disable=unused-variable
    return True

def too_many_arguments(a, b, c, d, e, f, g, h, i, j, k):
    """Function with too many arguments"""
    return a + b + c + d + e + f + g + h + i + j + k

def undefined_variable():
    """Function with undefined variable"""
    print(undefined_var)  # pylint: disable=undefined-variable

def bad_variable_name():
    """Function with bad variable name"""
    A = 1  # pylint: disable=invalid-name
    return A

if __name__ == "__main__":
    unused_variable()
    too_many_arguments(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11)
    undefined_variable()
    bad_variable_name() 