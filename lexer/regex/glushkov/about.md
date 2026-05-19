# About Glushkov's Construction Algorithm

Glushkov's algorithm (a.k.a. The algorithm of Berry-Sethi) transforms a regular expression into
a nondeterministic finite automaton (NFA). The properties of the resulting NFA are different from
the one produced by Thompson's construction. Most notably, the NFA is free from epsilon transitions.
In addition, for every input character, a state can transition to multiple other states. In Thompson's
construction, the resulting NFA is built in a way that a state can have multiple E-transitions, but a
character transition only leads to one next state.
