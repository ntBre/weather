#!/usr/bin/gnuplot --persist

set terminal qt
set bmargin 2.5
set xdata time
set timefmt "%m-%d/%H:%M"
set ylabel "Temperature (°F)"
set y2label "Rain chance (%)" rotate by 270
set y2tics
#set yrange [%d:%d]
plot "/tmp/weather.dat" u 1:2 w linespoints lc rgb "red" title "Hourly", \
     "/tmp/weather.dat" u 1:3 w linespoints lc rgb "green" title "Dew Point", \
     "/tmp/weather.dat" u 1:($4 == -999 ? NaN : $4) w linespoints lc rgb "blue" title "Wind Chill", \
     "/tmp/weather.dat" u 1:5 w boxes lc rgb "cyan" title "Rain" axes x1y2
