

USES Crt,Graph;

var Gd, Gm: INTEGER;
    KEY : CHAR;
    NUMBER : INTEGER;
    DOWN,COUNT,ACROSS : INTEGER;


PROCEDURE GRAPHIC; forward;
PROCEDURE CHERRY(ACROSS,DOWN:INTEGER); forward;
PROCEDURE PLUM(ACROSS,DOWN:INTEGER); forward;
PROCEDURE BAR(ACROSS,DOWN:INTEGER);  forward;
PROCEDURE DOLLAR(ACROSS,DOWN:INTEGER); forward;
PROCEDURE VIEW_PICS; forward;
PROCEDURE REEL; forward;
PROCEDURE SPIN; forward;
PROCEDURE BORDER; forward;
PROCEDURE WIN; forward;
PROCEDURE PLAY; forward;
PROCEDURE INSTRUCTIONS; FORWARD;
PROCEDURE MENU; forward;
procedure question; forward;
procedure exitgame; forward;
PROCEDURE GRAPHIC;
BEGIN
     Gd := Detect;
     InitGraph(Gd, Gm, 'C:\tp7\bgi');
     if GraphResult <> grok then Halt(1);


END;


{*****CHERRY*****}

PROCEDURE CHERRY(ACROSS,DOWN:INTEGER);

begin

{*****CIRCLE 1*****}
     SETCOLOR(RED);
     Circle(ACROSS, DOWN, 10);
     SETFILLSTYLE(SOLIDFILL,RED);
     FLOODFILL(ACROSS,DOWN,RED);
{*****END*****}

{*****CIRCLE 2*****}
     SETCOLOR(RED);
     Circle(ACROSS+10, DOWN, 10);
     SETFILLSTYLE(SOLIDFILL,RED);
     FLOODFILL(ACROSS+15,DOWN,RED);
{*****END*****}

{*****CIRCLE 3*****}
     SETCOLOR(RED);
     Circle(ACROSS+5, DOWN+3, 10);
     SETFILLSTYLE(SOLIDFILL,RED);
     FLOODFILL(ACROSS+5,DOWN+10,RED);
{*****END*****}

{*****STEM*****}
     SETCOLOR(BROWN);
     RECTANGLE(ACROSS+4,DOWN-10,ACROSS+7,DOWN-25);
     SETFILLSTYLE(SOLIDFILL,BROWN);
     FLOODFILL(ACROSS+5,DOWN-11,BROWN);
{*****END*****}
end;
{*****END CHERRY*****}

{*****PLUM*****}


PROCEDURE PLUM(ACROSS,DOWN:INTEGER);

begin

{*****CIRCLE*****}
     SETCOLOR(BLUE);
     Circle(ACROSS, DOWN, 20);
     SETFILLSTYLE(SOLIDFILL,BLUE);
     FLOODFILL(ACROSS,DOWN,BLUE);
{*****END*****}

{*****STEM*****}
     SETCOLOR(BROWN);
     LINE(ACROSS+20,DOWN,ACROSS+23,DOWN+10);
{*****END*****}
end;
{*****END PLUM*****}

{*****BAR*****}


PROCEDURE BAR(ACROSS,DOWN:INTEGER);

BEGIN
{******RECTANGLE*****}
     SETCOLOR(YELLOW);
     RECTANGLE(ACROSS+70,DOWN,ACROSS-20,DOWN-50);
{*****END*****}

{*****WRITING*****}
     SETFILLSTYLE(SOLIDFILL,YELLOW);
     FLOODFILL(ACROSS+2,DOWN-3,YELLOW);
     SETCOLOR(RED);
     SETTEXTSTYLE(2,HORIZDIR,10);
     OutTextXY(ACROSS-6, DOWN-50, 'BAR');
{*****END*****}

END;
{*****END BAR*****}

{*****DOLLAR*****}


PROCEDURE DOLLAR(ACROSS,DOWN:INTEGER);

BEGIN
{*****'S' SHAPE*****}
     SETCOLOR(YELLOW);
     LINE(ACROSS,DOWN,ACROSS+35,DOWN);
     LINE(ACROSS,DOWN,ACROSS,DOWN+30);
     LINE(ACROSS+35,DOWN,ACROSS+35,DOWN+5);
     LINE(ACROSS,DOWN+30,ACROSS+35,DOWN+30);
     LINE(ACROSS+35,DOWN+30,ACROSS+35,DOWN+60);
     LINE(ACROSS,DOWN+60,ACROSS+35,DOWN+60);
     LINE(ACROSS,DOWN+60,ACROSS,DOWN+55);
{*****END*****}

{*****DOWN LINES*****}
     LINE(ACROSS+12,DOWN-10,ACROSS+12,DOWN+70);
     LINE(ACROSS+25,DOWN-10,ACROSS+25,DOWN+70);
{*****END*****}
END;
{*****END DOLLAR*****}

{*****VIEW PICS*****}

PROCEDURE VIEW_PICS;

VAR
   KEY : CHAR;
BEGIN
     CHERRY(100,120);
     CHERRY(250,120);
     CHERRY(400,120);

     PLUM(100,200);
     PLUM(250,200);
     PLUM(400,200);

     BAR(100,300);
     BAR(250,300);
     BAR(400,300);

     DOLLAR(100,350);
     DOLLAR(250,350);
     DOLLAR(400,350);

END;

{*****END VIEW PICS*****}


PROCEDURE REEL;

BEGIN
     SETCOLOR(MAGENTA);
     LINE(100,100,100,400);
     LINE(100,100,500,100);
     LINE(500,100,500,400);
     LINE(100,400,500,400);

     LINE(215,100,215,400);
     LINE(350,100,350,400);


     cherry(150,140);
     plum(155,250);
     bar(135,380);

     DOLLAR(260,120);
     cherry(275,255);
     PLUM(280,355);

     BAR(400,160);
     DOLLAR(400,230);
     CHERRY(415,360);



END;


PROCEDURE SPIN;

VAR
   COUNT : INTEGER;

BEGIN
COUNT := 0;
REPEAT
INC(COUNT);
delay(250);
     SETFILLSTYLE(SOLIDFILL,BLACK);
     FLOODFILL(150,140,MAGENTA);
     FLOODFILL(250,160,MAGENTA);
     FLOODFILL(400,140,MAGENTA);

        DOLLAR(140,120);
        cherry(150,250);
        PLUM(155,355);

        BAR(250,160);
        DOLLAR(260,200);
        CHERRY(280,355);

        cherry(420,140);
        plum(425,250);
        bar(400,380);

delay(250);

     SETFILLSTYLE(SOLIDFILL,BLACK);
     FLOODFILL(150,140,MAGENTA);
     FLOODFILL(250,160,MAGENTA);
     FLOODFILL(400,140,MAGENTA);


     cherry(150,140);
     plum(155,250);
     bar(135,380);

     DOLLAR(260,120);
     cherry(275,255);
     PLUM(280,355);

     BAR(400,160);
     DOLLAR(400,230);
     CHERRY(415,360);
     delay(250);

     SETFILLSTYLE(SOLIDFILL,BLACK);
     FLOODFILL(150,140,MAGENTA);
     FLOODFILL(250,160,MAGENTA);
     FLOODFILL(400,140,MAGENTA);

     BAR(135,160);
     DOLLAR(145,220);
     CHERRY(155,360);

     cherry(270,140);
     plum(275,250);
     bar(260,380);

     DOLLAR(410,120);
     cherry(425,255);
     PLUM(430,355);

     delay(250);

UNTIL COUNT = 5;
     SETFILLSTYLE(SOLIDFILL,BLACK);
     FLOODFILL(150,140,MAGENTA);
     FLOODFILL(250,160,MAGENTA);
     FLOODFILL(400,140,MAGENTA);


END;

{*****BORDER*****}

PROCEDURE BORDER;

VAR
   ACROSS : INTEGER;

BEGIN
     SETCOLOR(GREEN);
     RECTANGLE(0,0,600,15);
     SETFILLSTYLE(SOLIDFILL,GREEN);
     FLOODFILL(1,1,GREEN);

     RECTANGLE(0,450,15,15);
     SETFILLSTYLE(SOLIDFILL,GREEN);
     FLOODFILL(1,35,GREEN);

     RECTANGLE(0,450,600,435);
     SETFILLSTYLE(SOLIDFILL,GREEN);
     FLOODFILL(445,445,GREEN);

     RECTANGLE(600,0,585,435);
     SETFILLSTYLE(SOLIDFILL,GREEN);
     FLOODFILL(597,35,GREEN);

END;

{*****win*****}

PROCEDURE WIN;

VAR
   RN : INTEGER;
   MONEY: INTEGER;

BEGIN
     RANDOMIZE;
     RN := (RANDOM(3)+1);

     IF RN = 1 THEN BEGIN
        CHERRY(150,240);
        CHERRY(270,240);
        CHERRY(425,240);
     END;

     IF RN = 2 THEN BEGIN
        PLUM(155,240);
        PLUM(275,240);
        PLUM(430,240);
     END;

     IF RN = 3 THEN BEGIN
        BAR(137,265);
        BAR(260,265);
        BAR(395,265);
     END;

     IF RN = 4 THEN BEGIN
        DOLLAR(145,205);
        DOLLAR(265,205);
        DOLLAR(425,205);
     END;

END;


{*****CREDITS *****}
PROCEDURE CREDITS;
VAR
   MONEY : INTEGER;

BEGIN
     CLOSEGRAPH;
     WRITELN(money);
     READLN;
     GRAPHIC;
     SETCOLOR(LIGHTBLUE);
     SETTEXTSTYLE(1,HORIZDIR,5);
     OutTextXY(130,15, 'POKER MACHINE');
     BORDER;
     REEL;
     SPIN;
END;

{*****PLAY*****}

PROCEDURE PLAY;

VAR
   ANS : CHAR;
   ST : INTEGER;
   WSPIN : INTEGER;
   MONEY : INTEGER;
   NUMBER : INTEGER;
   CHOICE : INTEGER;
label credits,continue;

BEGIN

     SETCOLOR(BLACK);
     SETFILLSTYLE(SOLIDFILL,BLACK);
     floodfill(100,100,GREEN);

     SETCOLOR(LIGHTBLUE);
     settextstyle(2, horizdir,7);
     OUTTEXTXY(30,60,'HOW MANY DOLLARS CAN YOU BET?');
     GOTOXY(5,8);
     READLN(MONEY);

     SETCOLOR(BLACK);
     SETFILLSTYLE(SOLIDFILL,BLACK);
     floodfill(100,100,GREEN);

     SETCOLOR(BLACK);
     SETTEXTSTYLE(1,HORIZDIR,5);
     OutTextXY(80,50, 'POKER MACHINE MENU');

     SETCOLOR(BLACK);
     SETFILLSTYLE(SOLIDFILL,BLACK);
     floodfill(100,100,GREEN);




     SETCOLOR(LIGHTBLUE);
     SETTEXTSTYLE(1,HORIZDIR,5);
     OutTextXY(130,15, 'POKER MACHINE');
     NUMBER := 10;
     SETCOLOR(lightRED);
     SETTEXTSTYLE(2,HORIZDIR,5);
     OutTextXY(150,76,'VIEW CREDITS');
     SETCOLOR(BLUE);
     SETTEXTSTYLE(2,HORIZDIR,5);
     OutTextXY(130,76,'1-');

     SETCOLOR(GREEN);
     RECTANGLE(120,93,260,76);
     SETFILLSTYLE(SOLIDFILL,GREEN);
     FLOODFILL(1,1,GREEN);


     REEL;
     ST := 0;
     RANDOMIZE;
     continue: REPEAT;
     MONEY := MONEY - 1;
     IF ST = 0 THEN WSPIN := RANDOM(10);
            SPIN;
            ST := ST + 1;
            IF ST <> WSPIN THEN REEL;
            IF ST = WSPIN THEN WIN;
            IF ST = WSPIN THEN ST := 0;
            Readln(choice);
            Case choice of
                 1 : goto CREDITS;
                 2 : PLAY;
                 {3 : author}
                 4 : exitgame;
            end;

UNTIL MONEY = 0;

credits:
     CLOSEGRAPH;
     WRITELN(money);
     READLN;
     GRAPHIC;
     SETCOLOR(LIGHTBLUE);
     SETTEXTSTYLE(1,HORIZDIR,5);
     OutTextXY(130,15, 'POKER MACHINE');
     BORDER;
     REEL;
     SPIN;

END;


{*****INSTRUCTIONS*****}

PROCEDURE INSTRUCTIONS;
var
   across,down :integer;

BEGIN
     CLOSEGRAPH;
     textcolor(yellow);
     for across := 2 to 78 do begin
          gotoxy(across,1);
          writeln('Í');
          gotoxy(across,20);
          writeln('Í');
     end;

     for down := 2 to 19 do begin
         gotoxy(1,down);
         writeln('º');
         gotoxy(79,down);
         writeln('º');
     end;

     textcolor(green);
     gotoxy(30,4);
     WRITELN('INSTRUCTIONS');
     READLN;
     GRAPHIC;
     border;
     menu;

END;

procedure question;
var
   money : integer;
begin
     play;
end;




{*****MENU*****}

PROCEDURE MENU;

VAR
   CHOICE : INTEGER;
   MONEY : INTEGER;
BEGIN
     SETCOLOR(YELLOW);
     SETTEXTSTYLE(1,HORIZDIR,5);
     OutTextXY(80,50, 'POKER MACHINE MENU');

     SETCOLOR(WHITE);
     SETTEXTSTYLE(2,HORIZDIR,8);
     OutTextXY(200,140, '1.');
     OutTextXY(200,170, '2.');
     OutTextXY(200,200, '3.');
     OutTextXY(200,230, '4.');
     SETCOLOR(LIGHTRED);
     SETTEXTSTYLE(2,HORIZDIR,8);
     OutTextXY(225,140, ' HOW TO PLAY GAME');
     OutTextXY(225,170, ' PLAY GAME');
     OutTextXY(225,200, ' CONTACTING AUTHOR');
     OutTextXY(225,230, ' EXIT GAME');
     SETCOLOR(LIGHTBLUE);
     OutTextXY(225,290, 'MAKE YOU CHOICE? ');
     GOTOXY(60,20);
     READLN(CHOICE);
     case choice of
          1 : instructions;
          2 : PLAY;
          {3 : author}
          4 : exitgame;
     end;

END;

procedure exitgame;
var
   sounds : integer;
begin
     border;
     setcolor(black);
     setfillstyle(solidfill,black);
     floodfill(100,100,green);

     setcolor(lightblue);
     settextstyle(1,horizdir,5);
     outtextxy(90,95,'Thanks for playing');
     sounds := 0;
     repeat
           inc(sounds);
           sound(random(4000)+200);
           delay(20);
     until sounds = 150;
     nosound;

end;



{*****MAIN*****}
BEGIN
     CLRSCR;
     GRAPHIC;
     BORDER;
     MENU;
{PLAY;}

     READLN;
     CLOSEGRAPH;

END.





