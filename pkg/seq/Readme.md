# To Do

## minimum length for sequences when writing
Sometimes we manipulate sequences and they become silly short - just a few bases/residues. Add an option to WriteToF() for a minimum length. Do not write a sequence if it is shorter than this.
This will break the interface a bit. s_opts will get another member. This means that zero has to mean "do not enforce a minimum length".
