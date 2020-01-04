# 2 nov 2018
# have a look at the effect of buffering the channel and the slice size
# used for gathering atom names
df = read.table (file = 'to_plot.dat')

colnames(df)[2] = 'sl_size'
colnames(df)[6] = 'time'
colnames(df)[10] = 'rss' # resident set size
#df = data.frame(sl_size=df$sl_size, time=df$time, rss = df$rss)
df$V1=NULL
df$V3=NULL
df$V5=NULL
df$V7=NULL
#a = df[df$chan_size == 0]
plot(x=df$sl_size, y=df$rss)
