script=`dirname $0`
projectsdir=`realpath $script/../projects`

for p in `ls $projectsdir | egrep -v "BKP|BAK|LOCK|[.]"`; do
    echo "PROJECT: $p"
    nSess=0
    for s in `find $projectsdir/$p -type d | egrep -v "BKP|BAK|LOCK|[.]" | sed "s|$projectsdir/||g" | egrep "^[^/]+/[^/]+/[^/]+$" | sed "s|^$p/||"`; do
	nSess=$((nSess + 1));
	wavs=`ls $projectsdir/$p/$s | egrep -c "[.]wav$"`
	wavSinPlu="wav"
	if [ $wavs -ne 1 ]; then
	    wavSinPlu="${wavSinPlu}s"
	fi
	locked=`ls -a $projectsdir/$p/$s | egrep -c "[.]lock$" | sed 's/^1$/*/' | sed 's/^0//g'`
	echo "$p/$s$locked	$wavs $wavSinPlu"
    done
    sessionSinPlu="session"
    if [ $nSess -ne 1 ]; then
       sessionSinPlu="${sessionSinPlu}s"
    fi
    echo "==>  $nSess $sessionSinPlu"
    echo ""
done
