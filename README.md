# Intro
The name will change.
This is a small set of packages for working with conservation in multiple sequence alignments.

# kl
Given two files, calculate the per-site Kullbach-Leibler distance, as well as the cosine similarity.

# entropy
Calculate the per-site entropy in a multiple sequence alignment.

# Structure
There are a few commands. These live under `cmd`. They are very short and quickly drop into a package which lives under `pkg`.

# To Do

Add an option for chimera output.

# Regrets

## Precision
This is not a regret. All the calculations are done in single precision. That is fine and accurate enough for us. It does mean that the code is full of `float32` casts.
