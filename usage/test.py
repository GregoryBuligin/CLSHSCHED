#!/usr/bin/env python3
import random

def recur_fibo(n):
    if n <= 1:
        return n
    else:
        return(recur_fibo(n-1) + recur_fibo(n-2))

if __name__ == "__main__":
    print(recur_fibo(random.randint(1, 35)))
