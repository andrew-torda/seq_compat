# To Do

## Speed up writing of sequences
If we are writing sequences, we can do it in the background while thinking about the .csv file.
In writeLen(), we should 

 - check if outseqFname is ""
 - set up a channel for errors
 - call seq.WriteToF() as a goroutine
 
