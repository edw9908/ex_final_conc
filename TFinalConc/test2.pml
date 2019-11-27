#define wait(s) atomic { s > 0 -> s-- }
#define signal(s) s++

byte S = 1
byte test = 0
byte turn = 0

active proctype P() {
    do
    ::
        if
        :: (turn == 0) ->
            test++
            printf("1\n")
            printf("2\n")
            printf("3\n")
            assert(test < 2)
            test--
            turn = 1
        fi
    od
}

active proctype Q() {
    do
    ::
        if
        :: (turn == 1) ->
            test++
            printf("4\n")
            printf("5\n")
            printf("6\n")
            assert(test < 2)
            test--
            turn = 0
        fi
    od
}