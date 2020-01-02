# easy-rv

## About
This project provides an easy-to-use implementation of _Runtime Verification_, based on formal semantics.

This project presents a more generalised any-type monitoring system, which can be used with any C project. 
_easy-rv_ was ported from [easy-rte](https://github.com/PRETgroup/easy-rte), which implemented a similar semantics for runtime enforcement.

## What is Runtime Verification?

Runtime Verification is a type of Run-time Assurance (RA) which focuses on formal semantics for monitoring the behaviour of systems.
As a system executes, the monitor observes according to formal policies. 
In each instant (a.k.a. tick/reaction), the monitor will output one of `{currently true, currently false, true, false}`, to indicate the status of the system.
This project aims to join the growing chorus calling for the use of Runtime Assurance mechanisms for use within the real-world cyber-physical systems domain. Using Runtime Verification, we can monitor the behaviour of _unpredictable_ and _untrustworthy_ processes, ensuring that they any violation of desired policies is detected.

## How does Runtime Verification work?

Runtime Verification works by placing a special observer execution module _between_ your plant and your controller.

```
+-------+                +------------+               
|       | ---inputs--->  |            |
| Plant |                | Controller |
|       | <--outputs---  |            |
+-------+                +------------+
```
becomes
```
             Policies
                \/
            +---------+
            |         |
            | Monitor |
            |         |
            +---------+
                 |
                 |
+-------+       /|       +------------+               
|       | ---inputs--->  |            |
| Plant |        /       | Controller |
|       | <--outputs---  |            |
+-------+                +------------+
```

In our case, we can compile our policies to either *C* or *Verilog* code. 

The C monitors are designed to be composed with your software in a system such as a microcontroller or an Arduino.

However, software monitors cannot by their nature monitor the behaviour of the hardware that they run upon. 
So, in this case you may want to build your monitor to Verilog, and then compose your microcontroller with custom hardware (such as on an FPGA or ASIC) to ensure system correctness.

## Build instructions

Download and install the latest version of [Go](https://golang.org/doc/install).

Then, download this repository, and run `make` or `make default`, which will generate the tools. 

* The ab5 example can be generated using `make c_mon PROJECT=ab5`.
* You can also generate Verilog monitors by using `make verilog_mon PROJECT=ab5`, for example. The Verilog monitors are a little trickier to use, and require I/O connections to be provided to them. They do not embed a function call to the controller inside.
* If you are interested in using the model checkers:
  * Obtain CBMC (C model checker) by running `sudo apt install cbmc`. Website here: https://www.cprover.org/cbmc/
  * Obtain EBMC (Verilog model checker) by going to website https://www.cprover.org/ebmc/

## A note on Easy-rv language

Easy-rv is based on Structured Text (ST) operators and syntax. When making guards, ensure that you adhere to the following operators:

| Operator       |     Code    |
| -------------- | ----------- |
| Assignment     |  `:=`       |
| Equality       |  `=`        |
| Inequality     |  `<>`       |
| Addition       |  `+`        |
| Subtraction    |  `-`        |
| Multiplication |  `*`        |
| Division       |  `/`        |
| Not            | `!` or NOT  |
| And            | `&&` or AND |
| Or             | `\|\|` or OR  |
| Brackets       | `(` and `)` |

## Example of Use (Pizza)

Let us consider the case of a frozen pizza. 
We define a food safety property via the following discrete timed automata:

![Pizza Lifespan Image](/example/pizza/Easy-RV-Pizza.png)

We can represent this state machine with the following _erv_ specification:

```
monitor pizza;
interface of pizza {
    uint32_t t;
}

policy FoodSafety of pizza {
    internals {
        dtimer_t xloc; //local timer (in minutes)
        dtimer_t xage; //total age of pizza (in minutes)

        constant uint32_t MAX_AGE := 4320;
    }

    states {

        //the pizza is frozen
        s_frozen rejecting {

            //the pizza is beginning to warm.
            -> s_warming on t > 0: xloc := 0, xage := 0;      

            //the pizza is still frozen.
            -> s_frozen on t <= 0;          
        }

        //the pizza is cooking
        s_warming rejecting {

            //the pizza is cooked too hot.
            -> s_burned on t >= 200;

            //it's taken too long to reheat the pizza.
            -> s_too_old on xloc > 100;

            //the pizza is cooked and now needs to be cooled.
            -> s_cooling on t >= 180 and t < 200 and xloc > 20: xloc := 0; 

            //the pizza is still cooking.
            -> s_warming on t < 180 and xloc < 100;
        }

        //the pizza is cooling after being cooked.
        s_cooling rejecting {
            
            //the pizza is ready to eat (notice we don't reset the timer.)
            -> s_ready on t < 50 and xloc <= 60 and xage <= MAX_AGE;

            //the pizza is too old (has been left out too long)
            -> s_too_old on xloc > 60 or xage > MAX_AGE;

            //the pizza is still too hot to eat
            -> s_cooling on t >= 50 and xloc <= 60 and xage <= MAX_AGE;
        }

        //the pizza is ready to eat
        s_ready accepting {

            //the pizza is still ready to eat
            -> s_ready on t >= 6 and xloc <= 60 and xage <= MAX_AGE;

            //the pizza is being refridgerated after cooking
            -> s_fridge on t < 6 and xloc <= 60 and xage <= MAX_AGE: xloc := 0;

            //the pizza is too old (has been left out too long)
            -> s_too_old on xloc > 60 or xage > MAX_AGE;
        }

        //the pizza is in the fridge and can still be eaten (some people like cold pizza)
        s_fridge accepting {

            //the pizza is still in the fridge (no more than 3 days)
            -> s_fridge on t <= 6 and xage <= MAX_AGE;

            //the pizza is being reheated
            -> s_reheating on t >= 6 and xage <= MAX_AGE: xloc := 0;

            //the pizza is too old (has been left in the fridge too long)
            -> s_too_old on xage > MAX_AGE;
        }

        //the pizza is being reheated
        s_reheating rejecting {

            //the pizza is still being reheated
            -> s_reheating on t < 100 and xloc < 100;

            //the pizza is ready to be cooled
            -> s_cooling on t > 100 and t < 200 and xloc <= 100: xloc := 0;

            //it's taken too long to reheat the pizza.
            -> s_too_old on xloc > 100;

            //the pizza is burned
            -> s_burned on t >= 200;
        }

        //the pizza is burned
        s_burned rejecting trap; //no way to recover this.

        //the pizza is too old
        s_too_old rejecting trap; //no way to recover this.
    }
}

```

As can be seen, this can be thought of as a simple mealy finite state machine, which provides the rules for correct operation.

We may now compile this system to executable code.

With _easy-rv_, this process is completed automatically in two steps. Firstly, we convert the _erv_ file into an equivalent policy XML file (which makes it easier to understand, and allows portability between tools).
* `./easy-rv-parser  -i example/pizza/pizza.erv -o example/pizza/pizza.xml`

Then, we convert this policy XML file into executable code, which is written in C. 
* `./easy-rv-c -i example/pizza/pizza.xml -o example/pizza`

Now, we can provide a `main.c` file which has our controller and plant interface code in it, and then compile the project together. In our case this is called `pizza_main.c`, and provides an example trace of temperatures (and will print the status of the monitor):

```c
//pizza_main.c contents:
#include "F_pizza.h"
#include <stdio.h>
#include <stdint.h>

void print_data(uint32_t count, monitorvars_pizza_t mon, io_pizza_t io) {
    printf("Tick %6d: temp:%5d C, STATE: (%4d, %4ld, %4ld), STATUS (Can Eat?):", count, io.t, mon._policy_FoodSafety_state, mon.xloc, mon.xage);
    switch(pizza_check_rv_status_FoodSafety(&mon)) {
        case 0: printf("TRUE"); break;
        case 1: printf("CURRENTLY TRUE"); break;
        case 2: printf("CURRENTLY FALSE"); break;
        default: printf("FALSE"); break;
    }
    printf("\r\n");
}

int main() {
    monitorvars_pizza_t mon;
    io_pizza_t io;
    
    pizza_init_all_vars(&mon, &io);

    uint32_t count = 0;
    io.t = 0;
    while(count++ < 100) {
        if(count == 2) io.t = 5; //start heating the pizza
        if(count == 5) io.t = 10;
        if(count == 15) io.t = 30;
        if(count == 20) io.t = 60;
        if(count == 25) io.t = 180; //pizza fully heated, remove from oven
        if(count == 30) io.t = 140;
        if(count == 35) io.t = 40; //pizza is cooled enough to eat

        //wait here until pizza expires

        pizza_run_via_monitor(&mon, &io);

        print_data(count, mon, io);
    }
}

void pizza_run(io_pizza_t *io) {
    //nothing to do, the pizza has no operations
}
```

If we run this example, it will present us an example trace:

```
Tick      1: temp:    0 C, STATE: (   0,    1,    1), STATUS (Can Eat?):CURRENTLY FALSE
Tick      2: temp:    5 C, STATE: (   1,    0,    0), STATUS (Can Eat?):CURRENTLY FALSE
Tick      3: temp:    5 C, STATE: (   1,    1,    1), STATUS (Can Eat?):CURRENTLY FALSE
.....
Tick     22: temp:   60 C, STATE: (   1,   20,   20), STATUS (Can Eat?):CURRENTLY FALSE
Tick     23: temp:   60 C, STATE: (   1,   21,   21), STATUS (Can Eat?):CURRENTLY FALSE
Tick     24: temp:   60 C, STATE: (   1,   22,   22), STATUS (Can Eat?):CURRENTLY FALSE
Tick     25: temp:  180 C, STATE: (   2,    0,   23), STATUS (Can Eat?):CURRENTLY FALSE
Tick     26: temp:  180 C, STATE: (   2,    1,   24), STATUS (Can Eat?):CURRENTLY FALSE
.....
Tick     33: temp:  140 C, STATE: (   2,    8,   31), STATUS (Can Eat?):CURRENTLY FALSE
Tick     34: temp:  140 C, STATE: (   2,    9,   32), STATUS (Can Eat?):CURRENTLY FALSE
Tick     35: temp:   40 C, STATE: (   3,   10,   33), STATUS (Can Eat?):CURRENTLY TRUE
Tick     36: temp:   40 C, STATE: (   3,   11,   34), STATUS (Can Eat?):CURRENTLY TRUE
....
Tick     84: temp:   40 C, STATE: (   3,   59,   82), STATUS (Can Eat?):CURRENTLY TRUE
Tick     85: temp:   40 C, STATE: (   3,   60,   83), STATUS (Can Eat?):CURRENTLY TRUE
Tick     86: temp:   40 C, STATE: (   7,   61,   84), STATUS (Can Eat?):FALSE
Tick     87: temp:   40 C, STATE: (   7,   62,   85), STATUS (Can Eat?):FALSE
....

```

To compile it together with the example main file, run `make c_mon c_build PROJECT=pizza`

## Example of Use (AB5)

Imagine a function which inputs boolean `A` and outputs boolean `B`. 
In _easy-rv_, we present this function with the following _erv_ syntax:
```
monitor ab5;
interface of ab5 {
	bool A;  //in this case, A goes from PLANT to CONTROLLER
	bool B; //in this case, B goes from CONTROLLER to PLANT
}
```

This is equivalent to the following C code, which is autogenerated, so that the function ab5Function_run can be provided by the user:
```c
//IO of the monitor ab5Function
typedef struct {
	bool A;
} io_ab5_t;


void ab5_run(io_ab5_t io);
```

Let's now give our function the following I/O properties:
1. A and B alternate starting with an A. 
2. B should be true within 5 ticks after an occurance of A.
3. A and B can only happen simultaneously in the first instant, or after any B. This indicates the end of a run.

We can present this as the following _easy-rv_ policy format:
```
policy AB5 of ab5 {
	internals {
		dtimer_t v;
	}

	states {

		//first state is initial, and represents "We're waiting for an A"
		s0 accepting {
			//if we receive neither A nor B, do nothing														
			-> s0 on (!A and !B): v := 0;

			//if we receive an A only, head to state s1							
			-> s1 on (A and !B): v := 0;
			
			//if we receive a !A and B then VIOLATION							
			-> violation on (!A and B);

			//if we receive both A and B then done
			-> done on (A and B);	
		}

		//s1 is "we're waiting for a B, and it needs to get here within 5 ticks"
		s1 rejecting {
			//if we receive nothing, and we aren't over-time, then we do nothing														
			-> s1 on (!A and !B and v < 5);	

			//if we receive a B only, head to state s0					
			-> s0 on (!A and B);

			//if we go overtime, or we receive another A, then VIOLATION	
			-> violation on ((v >= 5) or (A and B) or (A and !B));
		}

		done accepting trap;

		violation rejecting trap;
	}
}
```



