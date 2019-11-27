#define wait(s) atomic { s > 0 -> s-- }
#define signal(s) s++

byte S = 1
byte test = 0

active proctype P() {
    do
    ::
        wait(S)
        test++
        printf("Proceso P\n")
        printf("Proceso P\n")
        printf("Proceso P\n")
        assert(test < 2)
        test--
        signal(S)
    od
}

active proctype Q() {
    do
    ::
        wait(S)
        test++
        printf("Proceso Q\n")
        printf("Proceso Q\n")
        printf("Proceso Q\n")
        assert(test < 2)
        test--
        signal(S)
    od
}