import time

from py_beta.app import App

app = App()

t1 = time.time()
app.browse(1, 100)
t2 = time.time()
print(t2 - t1)

t3 = time.time()
app.browse(100, 100)
t4 = time.time()
print(t4 - t3)