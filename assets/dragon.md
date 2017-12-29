
The [Dragon curve](https://en.wikipedia.org/wiki/Dragon_curve) can be described using the [Lindenmayer system](https://en.wikipedia.org/wiki/L-system):

```
Rules:
	X ← X+YF+
	Y ← -FX-Y
Axiom:
	FX
Level:
	12
```

Here X and Y are variables and +, - and F are the turn right, turn left and go forward controls.

The following picture is made with [turtle](https://github.com/aicodix/turtle):

![Dragon curve](dragon.png)

[back to the top](/)

