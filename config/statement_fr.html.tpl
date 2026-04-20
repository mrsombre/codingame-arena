<!-- LEAGUES level1 -->

<div id="statement_back" class="statement_back" style="display: none"></div>
<div class="statement-body">
  <!-- LEAGUE ALERT -->
   <!-- 
  <div
    style="
      color: #7cc576;
      background-color: rgba(124, 197, 118, 0.1);
      padding: 20px;
      margin-right: 15px;
      margin-left: 15px;
      margin-bottom: 10px;
      text-align: left;
    "
  >
    <div style="text-align: center; margin-bottom: 6px">
      <img
        src="//cdn.codingame.com/smash-the-code/statement/league_wood_04.png"
      />
    </div>
    <p style="text-align: center; font-weight: 700; margin-bottom: 6px">
      This is a <b>league based</b> challenge.
    </p>
    <span class="statement-league-alert-content">
      For this challenge, multiple leagues for the same game are available. Once
      you have proven your skills against the a Boss, you will access a
      higher league and harder opponents will be available.
      <br /><br />
    </span>
  </div>
-->

  <!-- GOAL -->
  <div class="statement-section statement-goal">
    <h2>
      <span class="icon icon-goal">&nbsp;</span>
      <span>Goal</span>
    </h2>
    <div class="statement-goal-content">
      <div>
        Récupérez de l'énergie pour faire grandir vos robot-serpents et assurez-vous d'avoir les plus long robots en fin de jeu.
      </div>

      <!-- <div style="text-align: center; margin: 15px">
        <img src="https://static.codingame.com/servlet/fileservlet?id=1" style="width: 60%; max-width: 300px" />
      </div> -->
    </div>

  </div>


  <!-- RULES -->
  <div class="statement-section statement-rules">
    <h2>
      <span class="icon icon-rules">&nbsp;</span>
      <span>Règles</span>
    </h2>

    <div class="statement-rules-content">
      <p>
        Le jeu se joue sur une <b>grille</b>.
      </p>
      <p>
        Chaque joueur contrôle une équipe de <b>snakebots</b>. À chaque tour, les snakebots se déplacent <b>simultanément</b> selon les commandes des joueurs.
      </p>


      <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        🗺️ Carte
      </h3>
      <p>
        La grille est vue de côté et est composée de <b>plateformes</b>. Les plateformes sont des cases infranchissables.
      </p>
      
      <p>
        Chaque case de la grille peut contenir du vide, une plateforme, un segment du <b>corps</b> d’un snakebot ou de l’<b>énergie</b>.
      
      <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        🐍 Snakebot
      </h3>
      <p>
        Les snakebots sont constitués de plusieurs parties de corps adjacentes occupant chacune une cellule. La première cellule est leur <b>tête</b>.
      </p>
      <p>
        Les snakebots sont affectés par la <b>gravité</b>. Cela signifie qu’<b>au moins une partie du corps</b> doit se trouver au-dessus de quelque chose de <b>solide</b>, sinon ils vont <b>tomber</b>.
      </p>

      <p>
        Les autres snakebots sont considérés comme <b>solides</b>, tout comme les <b>plateformes</b> et l’<b>énergie
      </b>.

            <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        ↪️ Déplacement
      </h3>

      <p>
        Les snakebots sont <b>en mouvement permanent</b> et, à chaque tour, continuent dans la dernière direction qu’ils suivaient, sauf si le joueur leur donne une nouvelle direction.
      </p>
      <p>La direction initiale est <b>vers le haut</b>.

      <p>
        Lors d’un déplacement, un snakebot avance sa tête dans la direction à laquelle il fait face, et le reste de son corps le suit.
      </p>

      <p>
        <b>Cas 1 :</b> Si la cellule dans laquelle la tête du snakebot s’est déplacée contient une <b>plateforme</b> ou une autre <b>partie de corps</b>, la tête du snakebot est détruite et la partie suivante de son corps devient la <b>nouvelle tête</b>. Cela ne se produit que s’il reste au moins trois parties. Sinon, le snakebot entier est supprimé.
      </p>

      <p>
        <b>Cas 2 :</b> Si la cellule dans laquelle le snakebot s’est déplacé contient de l’<b>énergie</b>, le snakebot va <b>la manger</b>. Cela a deux effets :
        <ol>
          <li>Le snakebot <b>grandit</b> : une nouvelle partie du corps apparaît à l’extrémité de son corps.</li>
          <li>L'énergie disparait et cette cellule n’est plus considérée comme <b>solide</b>.</li>
        </ol>

    </p>
  <p>
    Ces collisions sont résolues simultanément pour tous les snakebots.
    </p>
    <p>
      <b>Cas spécial :</b> Si plusieurs têtes de snakebots entrent en collision sur une cellule contenant de l’<b>énergie</b>, cette énergie est considérée comme <b>mangée par chacun des snakebots</b> !
    </p>

      <p>
        Une fois les déplacements et suppressions résolus, tous les snakebots <b>tombent</b> vers le bas jusqu’à ce qu’une partie de leur corps repose sur quelque chose de <b>solide</b>.
      </p>

      <p>
        Il est possible d’étendre votre <b>snakebot</b> au-delà des limites de la grille. Mais tomber hors de la zone de jeu entraînera la suppression du snakebot.
      </p>
        
       
      <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        🎬 Actions
      </h3>
      <p>À chaque tour, les joueurs doivent fournir au moins une action sur la sortie standard.</p>
      <p>
        Les actions doivent être séparées par un point-virgule 
        <action>;</action>
        et être l’une des suivantes.
      </p>

      Commandes pour déplacer le snakebot d’identifiant <var>id</var> :

      <ul style="margin-bottom: 0">
        <li>
          <action>id UP</action> : donne la direction UP au snakebot <const>(x,y)=(0,-1)</const>.
        </li>
        <li>
          <action>id DOWN</action> : donne la direction DOWN au snakebot <const>(x,y)=(0,1)</const>.
        </li>
        <li>
          <action>id LEFT</action> : donne la direction LEFT au snakebot <const>(x,y)=(-1,0)</const>.
        </li>
        <li>
          <action>id RIGHT</action> : donne la direction RIGHT au snakebot <const>(x,y)=(1,0)</const>.
        </li>
      </ul>
      <p>
      Toute action de déplacement peut être suivie d’un texte qui sera affiché au-dessus du snakebot concerné à des fins de <b>débogage</b>.
      </p>
      <p>
        Commandes spéciales :
      </p>
      <ul style="margin-bottom: 0">
        <li>
          <action>MARK x y</action> : place un marqueur aux coordonnées spécifiées. Les marqueurs sont visibles dans le viewer à des fins de <b>débogage</b>.
        </li>
        <li>
          <action>WAIT</action>
          : ne rien faire.
        </li>
      </ul>
      <br>
 <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        ⛔ Fin de partie
      </h3>
      La partie se termine si l’une des conditions suivantes est vraie à la fin d’un tour :
      <ul style="margin-bottom: 0">
        <li>Tous les snakebots d’un joueur ont été supprimés.</li>
        <li>Il ne reste plus d’<b>énergie</b> à manger.</li>
        <li>200 tours se sont écoulés.</li>
      </ul>

      <br>
      <!-- Victory conditions -->
      <div class="statement-victory-conditions">
        <div class="icon victory"></div>
        <div class="blk">
          <div class="title">Conditions de victoire</div>
          <div class="text">
              Avoir plus de parties de corps au total sur l’ensemble de vos snakebots que votre adversaire à la fin de la partie.
          </div>
        </div>
      </div>
      <!-- Lose conditions -->
      <div class="statement-lose-conditions">
        <div class="icon lose"></div>
        <div class="blk">
          <div class="title">Conditions de défaite</div>
          <div class="text">
            Votre programme ne fournit pas de commande dans le temps imparti ou
            l’une des commandes est invalide.
          </div>
        </div>
      </div>
      <br />

      <!-- EXPERT RULES -->
      <div class="statement-section statement-expertrules">
        <h2>
          <span class="icon icon-expertrules">&nbsp;</span>
          <span>Détails techniques</span>
        </h2>
        <div class="statement-expert-rules-content">
          <ul style="padding-left: 20px;padding-bottom: 0">
            <li>
              Le code source de ce jeu sera publié sur <a rel="nofollow" target="_blank"
              href="https://github.com/CodinGame/WinterChallenge2026-Exotec">ce dépôt GitHub</a>.
              </li>
           
          </ul>
        </div>
      </div>

      <div class="statement-section statement-expertrules">
        <h2>
          <span>🐞 Conseils de débogage</span>
        </h2>
        <ul>
      <li>
        L'action <action>MARK x y</action> permet d'afficher jusqu'à 4 marqueurs par tour dans le viewer.
        </li>

          <li>
            Survolez la grille pour voir des informations supplémentaires sur la cellule sous votre curseur.
          </li>

          <li>
            Appuyez sur l’icône d’engrenage du viewer pour accéder à des options d’affichage supplémentaires.
          </li>
          <li>
            Utilisez le clavier pour contrôler l’action : espace pour lire/pause, flèches pour avancer image par image.
          </li>
        </ul>
      </div>
    </div>
  </div>




    <!-- PROTOCOL -->
  <details class="statement-section statement-protocol">
    
      <summary
        open
        style="cursor: pointer; margin-bottom: 10px; display: inline-block"
      >
        <span style="display: inline-block; margin-bottom: 10px"
          >Cliquer pour développer</span
        >
        <h2 style="margin-bottom: 0">
          <span class="icon icon-protocol">&nbsp;</span>
          <span>Protocole de jeu</span>
        </h2>
      </summary>

      <!-- Protocol block -->
      <div class="blk">
        <div class="title">Entrée d’initialisation</div>
        <div class="text">
          <span class="statement-lineno">Ligne 1 :</span> <var>myId</var>, un entier
 pour l’identification de votre joueur (<const>0</const> ou <const>1</const>).
            <br/>
            <span class="statement-lineno">Ligne 2 :</span> <var>width</var>, un entier correspondant à la largeur de la grille.
          <br/>
            <span class="statement-lineno">Ligne 3 :</span> <var>height</var> un entier correspondant à la hauteur de la grille.

          <br/>
          <br/>
            <span class="statement-lineno">Les <var>height</var> lignes suivantes </span>
            
            <span class="statement-lineno">:</span> Une ligne de la grille 
            contenant
            <var>width</var> caractères. Chaque caractère peut être :
          
          <ul>
            <li><const>#</const> : une <b>plateforme</b>.</li>
            <li><const>.</const> : une case <b>libre</b>.</li>
          </ul>
          
          <span class="statement-lineno">Ligne suivante :</span> <var>snakbotsPerPlayer</var> un entier correspondant au nombre de snakebots contrôlés par chaque joueur.
          <br/>
          <br/>
          <span class="statement-lineno">Les <var>snakbotsPerPlayer</var> lignes suivantes </span>
          
          <span class="statement-lineno">:</span> Un entier <var>snakebotId</var> pour chacun des snakebots contrôlés par <b>vous</b>.
          
          <br/><br/>
          <span class="statement-lineno">Les <var>snakbotsPerPlayer</var> Lignes suivantes </span>
          <span class="statement-lineno">:</span> Un entier <var>snakebotId</var> pour chacun des snakebots contrôlés par <b>votre adversaire</b>.
        </div>
      </div>
      <div class="blk">
        <div class="title">Entrée pour un tour de jeu</div>
        <div class="text">
          <p>
            <span class="statement-lineno">Première ligne :</span> <var>powerSourceCount</var>, un entier
             correspondant au nombre restant d’<b>énergie</b> sur la
            grille.
          </p>
          <p>
            <span class="statement-lineno">Les <var>powerSourceCount</var> lignes suivantes :</span> Les
            <const>2</const> entiers suivants pour chaque énergie :
          </p>
          <ul>
            <li><var>x</var> : coordonnée X</li>
            <li><var>y</var> : coordonnée Y</li>
          </ul>
          <p>
            <span class="statement-lineno">Ligne suivante :</span> <var>snakebotCount</var>, un entier
             correspondant au nombre restant de <b>snakebots</b> sur la
            grille.
          </p>
          <p>
            <span class="statement-lineno"> Les <var>snakebotCount</var> lignes suivantes </span>
            <span class="statement-lineno">:</span> Les
            <const>2</const> entrées suivantes pour chaque snakebot :
          </p>
          <ul>
            <li><var>snakebotId</var> : un entier correspondant à l’identifiant de ce snakebot</li>
            <li><var>body</var> : une chaîne de coordonnées séparées par des deux-points (<const>:</const>), chaque coordonnée représentant la position d’une partie du corps sous la forme « <const>x,y</const> ». La première coordonnée de la liste est la <b>tête</b>.<br>
            <em>Exemple : </em>« <const>0,1:1,1:2,1</const> »<em> correspond à un snakebot avec une tête en <const>x=0</const> et <const>y=1</const> et deux autres parties du corps vers la droite.</em></li>
          </ul>          
        </div>
      </div>

      <!-- Protocol block -->
      <div class="blk">
        <div class="title">Sortie</div>
        <div class="text">
          <p>
            Une seule ligne, contenant au moins une action :
          </p>
          <!-- <em>(Au moins une action, et pas plus de trente).</em> -->
          <ul>
            <li>
              <action>id UP</action> où id est le <var>snakebotId</var> d’un snakebot que vous contrôlez.
            </li>
            <li>
              <action>id DOWN</action> où id est le <var>snakebotId</var> d’un snakebot que vous contrôlez.
            </li>
            <li>
              <action>id LEFT</action> où id est le <var>snakebotId</var> d’un snakebot que vous contrôlez.
            </li>
            <li>
              <action>id RIGHT</action> où id est le <var>snakebotId</var> d’un snakebot que vous contrôlez.
            </li>
            <li>
              <action>MARK x y</action> : vous pouvez marquer jusqu’à <b>quatre</b> coordonnées par tour.
            </li>
            <li>
              <action>WAIT</action>
            </li>
          </ul>
          <p>
            Les instructions sont séparées par des points-virgules (<const>;</const>). Par exemple :
          </p>
          <p>
            <action><code>1 LEFT;2 RIGHT;MARK 12 2</code></action>
          </p>
        </div>
      </div>

      <div class="blk">
        <div class="title">Contraintes</div>
        <div class="text">
          Temps de réponse par tour ≤ <const>50</const>ms <br />Temps de réponse pour
          le premier tour ≤ <const>1000</const>ms
          <br />
          <const>15</const> &le; <var>width</var> &le; <const>45</const>
          <br />
          <const>10</const> &le; <var>height</var> &le; <const>30</const>
          <br />
          <const>1</const> &le; <var>snakebotCount</var> &le; <const>8</const>
        </div>
      </div>

    
    </details>

    

  <!-- SHOW_SAVE_PDF_BUTTON -->
</div>
