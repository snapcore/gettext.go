%{
package pluralforms
%}

%union {
	num int
	exp Expression
}

// This declares the same associativity and precedence as libintl's parser
%right '?'              /*   ?          */
%left '|'               /*   ||         */
%left '&'               /*   &&         */
%left '=' neTok            /*   == !=      */
%left ltTok gtTok lteTok gteTok
%left '+' '-'
%left '*' '/' '%'
%right '!'              /*   !          */

%token neTok ltTok gtTok lteTok gteTok
%token <num> numTok
%token invalidTok
%type <exp> exp

%%
start: exp
	{
		yylex.(*lexer).exp = $1
	}

exp: exp '?' exp ':' exp
	{
		$$ = ternaryExpr{test: $1, ifTrue: $3, ifFalse: $5}
	}
   | exp '|' exp
	{
		$$ = orExpr{left: $1, right: $3}
	}
   | exp '&' exp
	{
		$$ = andExpr{left: $1, right: $3}
	}
   | exp '=' exp
	{
		$$ = eqExpr{left: $1, right: $3}
	}
   | exp neTok exp
	{
		$$ = neExpr{left: $1, right: $3}
	}
   | exp ltTok exp
	{
		$$ = ltExpr{left: $1, right: $3}
	}
   | exp lteTok exp
	{
		$$ = lteExpr{left: $1, right: $3}
	}
   | exp gtTok exp
	{
		$$ = gtExpr{left: $1, right: $3}
	}
   | exp gteTok exp
	{
		$$ = gteExpr{left: $1, right: $3}
	}
   | exp '+' exp
	{
		$$ = addExpr{left: $1, right: $3}
	}
   | exp '-' exp
	{
		$$ = subExpr{left: $1, right: $3}
	}
   | exp '*' exp
	{
		$$ = mulExpr{left: $1, right: $3}
	}
   | exp '/' exp
	{
		$$ = divExpr{left: $1, right: $3}
	}
   | exp '%' exp
	{
		$$ = modExpr{left: $1, right: $3}
	}
   | '!' exp
	{
		$$ = notExpr{$2}
	}
   | 'n'
	{
		$$ = varExpr{}
	}
   | numTok
	{
		$$ = numberExpr{$1}
	}
   | '(' exp ')'
	{
		$$ = $2
	}
%%
