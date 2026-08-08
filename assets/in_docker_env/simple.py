from tqdm import tqdm
import time

for i in tqdm(range(20)):
    print(f"iter {i}")
    time.sleep(0.5)