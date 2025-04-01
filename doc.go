/*
Package mofu provides utilities to create a mock function to use in test code.

There are some internal queues in the mock object.
A queue is created for corresponding to an argument pattern.
Additionally there is a queue for the default pattern.

The mock function consumes an item on the top of the queue when the mock function is called.
If the queue is empty, the mock function returns default values or zero values.
*/
package mofu
