df = read.table (file='z')
colnames(df)[1] = 'nthread'
colnames(df)[3] = 'real'
colnames(df)[5] = 'user'
colnames(df)[7] = 'sys'

plot (x=df$nthread, y=df$real, ylim = c(0, 4100),
      xlab = 'n thread', ylab = 'time (s)', pch = 1, lwd=3)
points (x=df$nthread, y=df$user, xlab='', ylab= '', pch=2)
points (x=df$nthread, y=df$sys,  xlab='', ylab= '', pch=3)
legend ("topright", legend = c("real", "user", "sys"), pch=c(1, 2, 3))


