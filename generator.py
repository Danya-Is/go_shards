import random

if __name__ == '__main__':
    with open('ids.txt', 'w') as file:
        for i in range(10):
            file.write(f'{random.randint(0, 200000)}\n')