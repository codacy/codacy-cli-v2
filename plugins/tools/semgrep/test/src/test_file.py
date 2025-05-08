#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Test file for semgrep analysis
"""

import os
import sys
import subprocess

def unsafe_command_execution():
    """Function with unsafe command execution"""
    user_input = "ls -la"
    os.system(user_input)  # semgrep: python.lang.security.audit.subprocess-shell-true.subprocess-shell-true
    subprocess.run(user_input, shell=True)  # semgrep: python.lang.security.audit.subprocess-shell-true.subprocess-shell-true

def hardcoded_password():
    """Function with hardcoded password"""
    password = "mysecretpassword123"  # semgrep: python.lang.security.audit.hardcoded-password.hardcoded-password
    return password

def unsafe_deserialization():
    """Function with unsafe deserialization"""
    import pickle
    data = b"cos\nsystem\n(S'ls -la'\ntR."
    pickle.loads(data)  # semgrep: python.lang.security.audit.pickle.avoid-pickle

def main():
    unsafe_command_execution()
    hardcoded_password()
    unsafe_deserialization()

if __name__ == "__main__":
    main() 