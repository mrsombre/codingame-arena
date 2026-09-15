<!-- LEAGUES level1 level2 level3 -->
<div id="statement_back" class="statement_back" style="display: none"></div>
<div class="statement-body">
	<!-- LEAGUE ALERT -->
	<div style="
		color: #7cc576;
		background-color: rgba(124, 197, 118, 0.1);
		padding: 20px;
		margin-right: 15px;
		margin-left: 15px;
		margin-bottom: 10px;
		text-align: left;
		">
		<div style="text-align: center; margin-bottom: 6px">
			<img
				src="//cdn.codingame.com/smash-the-code/statement/league_wood_04.png"
				/>
		</div>
		<p style="text-align: center; font-weight: 700; margin-bottom: 6px">
			Ce challenge est basé sur un système de <b>ligues</b>.
		</p>
		<span class="statement-league-alert-content">
		Pour ce défi, plusieurs ligues pour le même jeu sont disponibles. Une fois
		vos compétences prouvées contre le premier Boss, vous accéderez à une
		ligue supérieure et des règles supplémentaires seront disponibles.
		<br /><br />
		Dans les premières ligues, votre soumission affrontera
		uniquement le boss dans l’arène. Gagnez au moins 3 fois sur 5 pour
		avancer.
		</span>
	</div>
	<!-- GOAL -->
	<div class="statement-section statement-goal">
		<h2>
			<span class="icon icon-goal">&nbsp;</span>
			<span>Objectif</span>
		</h2>
		<div class="statement-goal-content">
			<div>
				<!-- BEGIN level1 -->
				Les deux premières ligues ont un <b>objectif</b> spécial à atteindre. La partie complète se jouera,
				mais vous ne pouvez gagner qu’en remplissant l’objectif. Une fois fait, vous pourrez commencer à travailler sur votre bot complet.
				<!-- END -->
				<!-- BEGIN level3 -->
				<p>Marquez plus de points que votre adversaire en traçant des lignes de chemin de fer entre les villes tout en sabotant les efforts de vos adversaires.</p>
				<!-- END -->

				<!-- BEGIN level1 -->
				<h3 style="
					font-size: 16px;
					font-weight: 700;
					padding-top: 20px;
					color: #838891;
					padding-bottom: 15px;
					">
					🎯 Objectif de ligue 1 :
				</h3>
				Reliez deux <b>villes</b> pour former une <b>connexion ferroviaire active</b> et gagner instantanément la partie.<br><br>
				Votre adversaire <b>Boss</b> dans l'arène passera ses tours. Si vous ne parvenez pas à former une connexion en moins de 
				<const>100</const>
				tours, vous perdrez. Gagnez au moins 
				<const>3</const>
				fois sur 
				<const>5</const>
				pour progresser à la ligue suivante.
				<!-- END -->
				<!-- BEGIN level2 -->
				<h3 style="
					font-size: 16px;
					font-weight: 700;
					padding-top: 20px;
					color: #838891;
					padding-bottom: 15px;
					">
					🎯 Objectif de ligue 2 :
				</h3>
				Utilisez l’action 
				<action>DISRUPT</action>
				pour effacer au moins un rail de votre adversaire.<br>
				<br>
				L’IA <b>Boss</b> placera des rails entre des villes aléatoires. Pour gagner, perturbez une région contenant un rail ennemi suffisamment pour qu’elle soit <b>effacée à l’encre</b>, détruisant tout rail à l’intérieur. Gagnez au moins 
				<const>3</const>
				fois sur 
				<const>5</const>
				pour progresser à la ligue suivante. <em>Plus d’informations sur la perturbation plus bas.</em>
				<!-- END -->
			</div>
		</div>
	</div>
	<!-- RULES -->
	<div class="statement-section statement-rules">
		<h2>
			<span class="icon icon-rules">&nbsp;</span>
			<span>Règles</span>
		</h2>
		<div class="statement-rules-content">
			<p>Dans ce jeu, les deux joueurs utilisent de la <b>peinture</b> pour tracer des voies ferrées sur une carte <b>magique</b>. Relier les villes de la carte apportera la prospérité à votre propre monde.
			</p>
			<p>La carte est représentée dans le jeu par une <b>grille</b>.</p>
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
				La grille est composée de cases qui peuvent être de quatre types :
			<ul>
				<li>
					Type 
					<const>0</const>
					pour les <b>plaines</b>.
				</li>
				<li>
					Type 
					<const>1</const>
					pour les <b>rivières</b>.
				</li>
				<li>
					Type 
					<const>2</const>
					pour les <b>montagnes</b>.
				</li>
			</ul>
			</p>
			<p>
				La grille est divisée en <b>régions</b>. Chaque région est composée de plusieurs cases voisines. Chaque région possède un <var>regionId</var> (identifiant de région) unique. Les régions sont sensibles à la <b>perturbation</b> par les joueurs.
				<!-- BEGIN level1 -->
				<em>
				Plus d’informations sur la perturbation dans la ligue suivante.
				</em>
				<!-- END -->
				<!-- BEGIN level2 level3 -->
				<br><br/><em>Plus d’informations sur la perturbation plus bas.</em>
				<!-- END -->
			</p>
			<p>
				Certaines régions contiendront une <b>ville</b>. Les villes ne peuvent se trouver que sur des cases de <b>plaine</b>.
			</p>

			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🏯 Villes
			</h3>
			<p>
				Chaque partie commence avec plusieurs <b>villes</b> placées aléatoirement sur la carte. Il y en aura au maximum une par <b>région</b> et deux régions partageant une frontière ne contiendront jamais toutes deux une ville.
			</p>
			<p>
				Chaque ville possède un <var>townId</var> (identifiant de ville) unique.
			</p>
			<p>
				Chaque ville aura une liste de <var>desiredConnections</var> (connexions souhaitées) : une liste d’identifiants de villes représentant toutes les <b>autres villes</b> auxquelles cette ville souhaite être <b>connectée</b> via des <b>rail</b> placées par les joueurs.
			</p>
			<p>Fournir à une ville des <b>rails</b> la reliant à une ville souhaitée est la façon dont les joueurs marquent des <b>points</b>.
			</p>
			<p>
				Les connexions souhaitées sont <b>unilatérales</b>. Cela signifie que si la ville 
				<const>0</const>
				apparaît avec une connexion souhaitée vers la ville 
				<const>1</const>
				, la ville 
				<const>1</const>
				<b>ne voudra pas</b> se connecter à la ville
				<const>0</const>
				.
			</p>
			<p>
				Une ville peut avoir 
				<const>zero</const>
				<var>desiredConnections</var> (connexions souhaitées), mais sera toujours l’objet d’au moins une autre ville ayant des <var>desiredConnections</var>.
			</p>

			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🛤️ Placement des rails
			</h3>
			<p>
				À chaque tour, les joueurs reçoivent 
				<const>3</const>
				<b>points de peinture</b> qu’ils peuvent utiliser pour placer des <b>rails</b> sur la carte. ⚠️ Ces points ne se reportent pas au tour suivant et seront perdus s’ils ne sont pas utilisés.
			</p>
			<p>
				Coût :
			<ul>
				<li>
					<const>1</const>
					point de peinture pour placer un rail sur les <b>plaines</b>.
				</li>
				<li>
					<const>2</const>
					points de peinture pour placer un rail sur une <b>rivière</b>.
				</li>
				<li>
					<const>3</const>
					points de peinture pour placer un rail sur les <b>montagnes</b>.
				</li>
			</ul>
			</p>
			<p>
				Le <var>owner</var> (propriétaire) d’un rail est le <var>playerId</var> (
				<const>0</const>
				-
				<const>1</const>
				) du joueur qui l’a placée.
			</p>
			<p>
				Si les deux joueurs placent un rail au <b>même tour</b> et au <b>même endroit</b>, le <var>owner</var> du rail sera 
				<const>2</const>
				, indiquant un rail neutre.
			</p>
			<p>un rail ferrée <b>ne peut pas</b> être placée sur une <b>ville</b> ni sur un rail existante.</p>
			<p>
				Une fois placée, un rail ferrée se <b>connectera automatiquement</b> aux autres <b>rails</b> et <b>villes</b> adjacentes orthogonalement.
			</p>

			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🏯🛤️🏯 Connexions
			</h3>
			<p>
				Pour chaque paire de villes dont l'une possède l'autre dans ces <var>desiredConnections</var> (connexions souhaitée), s’il existe au moins un <b>chemin</b> entre les deux, le plus court de ces <b>chemins</b> devient la <b>connexion active</b> entre ces villes.
			</p>
			<p> Un <b>chemin</b> est une séquence ininterrompue de cases adjacentes orthogonalement contenant un <b>rail</b> ou une <b>ville</b>.
			</p>
			<p> S’il y a <b>plusieurs</b> chemins le plus court possible, le chemin choisi <b>privilégiera</b> toujours la direction dans lordre suivant lorsqu’on part de la ville demandeuse vers la ville souhaitée :
			<ol>
				<li>
					<const>NORTH</const>
				</li>
				<li>
					<const>EAST</const>
				</li>
				<li>
					<const>SOUTH</const>
				</li>
				<li>
					<const>WEST</const>
				</li>
			</ol>
			</p>
			<p>
				À la fin de chaque tour, chaque <b>connexion active</b> rapportera 
				<const>1</const>
				point à chaque joueur pour chaque rail qu’il <b>possède</b> dans le <b>chemin</b>.
			</p>
			<br>
			<p><strong>Exemple 1 :</strong></p>
			<div style="text-align: center; margin: 15px">
				<img src="https://static.codingame.com/servlet/fileservlet?id=151670109832251"
					style="width: 60%; max-width: 400px;" />
			</div>
			<p> Ici, il y a une <b>connexion active</b> de la ville <b>0</b> à la ville <b>1</b> et une autre de la ville <b>0</b> à la ville <b>2</b>.
			</p>
			<p>
				Les joueurs rouge et bleu gagneront tous deux 
				<const>3</const>
				points pour la connexion <b>0-1</b> et 
				<const>4</const>
				points pour la connexion <b>0-2</b> à la fin du tour.
			</p>
			<br/>
			<p><strong>Exemple 2 :</strong></p>
			<div style="text-align: center; margin: 15px">
				<img src="https://static.codingame.com/servlet/fileservlet?id=151670123783140"
					style="width: 60%; max-width: 400px" />
			</div>
			<p>
				Ici, seul le plus court chemin de <b>0</b> à <b>2</b> est utilisé pour la connexion <b>0-2</b>. Cela signifie que le joueur rouge gagnera 
				<const>3</const>
				points pour cette connexion, et le joueur bleu n’en gagnera aucun.
			</p>
			<p>
				Il existe deux chemins de même longueur entre <b>0</b> et <b>1</b>, mais comme 
				<const>EAST</const>
				a une priorité plus élevée que 
				<const>SOUTH</const>
				, le chemin choisi sera celui qui passe par la ville <b>2</b>. Le joueur rouge gagnera 
				<const>4</const>
				points pour cette connexion, et le joueur bleu n’en gagnera que 
				<const>1</const>.
			</p>

			<br>
			<!-- BEGIN level2 level3 -->
			<!-- BEGIN level2 -->
			<div class="statement-new-league-rule">
				<!-- END -->
				<h3 style="
					font-size: 16px;
					font-weight: 700;
					padding-top: 20px;
					color: #838891;
					padding-bottom: 15px;
					">
					💥 Perturbation
				</h3>
				<p>
					De façon similaire aux <b>points de peinture</b>, les joueurs peuvent aussi utiliser 
					<const>1</const>
					<b>point de perturbation</b>
					par tour. Il peut être utilisé pour altérer la carte, vous donnant un avantage sur votre adversaire.
					<em>Ces points ne sont pas conservés entre les tours.</em>
				</p>
				<p>
					Les joueurs peuvent dépenser leur <b>point de perturbation</b> à chaque tour pour augmenter l’<var>instability</var> (instabilité) de n’importe quelle <b>région</b> de 
					<const>1</const>
					.
				</p>
				<p>
					Une fois que l’<var>instability</var> (instabilité) d’une région atteint 
					<const>4</const>
					, cette région est <b>effacée à l’encre</b>, supprimant tout <b>rails</b> placées et rendant tout placement futur dessus <b>impossible</b>. Toute connexion active passant par cette région sera rompue.
				</p>
				<p>Il n’est pas possible de perturber une région qui est déjà <b>effacée à l’encre</b>.</p>
				<p>Les régions contenant une ville ne peuvent pas être perturbées.</p>
				<!-- BEGIN level2 -->
			</div>
			<!-- END -->
			<!-- END -->
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
				et doivent être l’une des suivantes :
			</p>
			<ul style="margin-bottom: 0">
				<li>
					<action>PLACE_TRACKS x y</action>
					: placer un rail sur une case libre.
				</li>
				<li>
					<action>AUTOPLACE fromX fromY toX toY</action>
					: génère automatiquement une liste d’actions pour le chemin <b>le moins coûteux</b> des coordonnées <var>from</var> au coordonnées <var>to</var> en termes de points de peinture.
					Cela ne fera rien si un chemin existe déjà.
					<br/>
					<em>Les actions générées remplacent cette commande.</em>
				</li>
			</ul>
			<!-- BEGIN level2 level3 -->
			<!-- BEGIN level2 -->
			<div class="statement-new-league-rule">
				<!-- END -->
				<ul style="margin-top: 0; margin-bottom: 0;">
					<li>
						<action>DISRUPT regionId</action>
						: augmente l’instabilité d’une région.
						<em>Remarque :</em>
						<action>DISRUPT x y</action>
						fonctionne aussi, pour cibler la région dont (x,y) fait partie.
					</li>
				</ul>
				<!-- BEGIN level2 -->
			</div>
			<!-- END -->
			<!-- END -->
			<ul style="margin-top: 0">
				<li>
					<action>WAIT</action>
					: ne rien faire.
				</li>
			</ul>
			<br>

			<!-- Victory conditions -->
			<div class="statement-victory-conditions">
				<div class="icon victory"></div>
				<div class="blk">
					<div class="title">Conditions de victoire</div>
					<div class="text">
						<ul>
							<!-- BEGIN level1 -->
							Former une <b>connexion active</b> et marquer n’importe quel nombre de points.
							<!-- END -->
							<!-- BEGIN level2 -->
							<li>
								Faire perdre à votre adversaire au moins un <b>rail</b> en faisant en sorte qu’une région soit <b>effacée à l’encre</b> en utilisant 
								<action>DISRUPT</action>
								.
							</li>
							<!-- END -->
							<!-- BEGIN level3 -->
							<li>
								Avoir le plus de points après 
								<const>100</const>
								tours.
							</li>
							<li>Être en tête si toutes les connexions souhaitées deviennent impossibles à réaliser.</li>
							<!-- END -->
						</ul>

					</div>
				</div>
			</div>
			<!-- Lose conditions -->
			<div class="statement-lose-conditions">
				<div class="icon lose"></div>
				<div class="blk">
					<div class="title">Conditions de défaite</div>
					<div class="text">
						<!-- BEGIN level1 level2 -->
						<ul>
							<li>
								Votre programme ne fournit pas de commande dans le temps imparti
								ou l’une des commandes est invalide.
							</li>
							<li>
								Vous ne complétez pas l’objectif en moins de 
								<const>100</const>
								tours.
							</li>
						</ul>
						<!-- END -->
						<!-- BEGIN level3 -->
						Votre programme ne fournit pas de commande dans le temps imparti
						ou l’une des commandes est invalide.
						<!-- END -->
					</div>

				</div>
			</div>
			<br />
			<!-- BEGIN level3 -->
			<!-- EXPERT RULES -->
			<div class="statement-section statement-expertrules">
				<h2>
					<span class="icon icon-expertrules">&nbsp;</span>
					<span>Détails techniques</span>
				</h2>
				<div class="statement-expert-rules-content">
					<ul style="padding-left: 20px;padding-bottom: 0">
						<li>
							Vous pouvez consulter le code source de ce jeu sur <a rel="nofollow" target="_blank"
							href="https://github.com/CGjupoulton/SummerChallenge2026">ce dépôt GitHub</a>.
						</li>
						<li>
							Toutes les actions 
							<action>PLACE_TRACKS</action>
							, y compris celles issues de 
							<action>AUTOPLACE</action>
							, sont effectuées <b>avant</b> les actions 
							<action>DISRUPT</action>.
							Les points sont comptés à la toute fin du tour,
							<b>après</b> l’encrage des régions instables.
						</li>
						<li>
							Les commandes correspondant à des actions impossibles sont ignorées. 
							Si une action impossible fait partie d’un 
							<action>AUTOPLACE</action>,
							le reste des actions générées est également ignoré, même si elles sont possibles.
						</li>
					</ul>
				</div>
			</div>

			<!-- END -->
			<div class="statement-section statement-expertrules">
				<h2>
					<span>🐞 Conseils de débogage</span>
				</h2>
				<ul>
					<li>
						Survolez la grille pour voir des informations supplémentaires sur la
						case sous votre souris.
					</li>
					<li>
						Appuyez sur l’icône d’engrenage du visualiseur pour accéder à des
						options d’affichage supplémentaires.
					</li>
					<li>
						Utilisez le clavier pour contrôler l’action : barre d’espace pour
						lire/mettre en pause, flèches pour avancer d’une image à la fois.
					</li>
				</ul>
			</div>
		</div>
	</div>
	<!-- PROTOCOL -->
	<!-- BEGIN level1 level2 -->
	<details class="statement-section statement-protocol">
		<!-- END -->
		<!-- BEGIN level3 -->
		<details open class="statement-section statement-protocol">
			<!-- END -->
			<summary open style="cursor: pointer; margin-bottom: 10px; display: inline-block">
				<span style="display: inline-block; margin-bottom: 10px"
					>Cliquez pour développer</span>
				<h2 style="margin-bottom: 0">
					<span class="icon icon-protocol">&nbsp;</span>
					<span>Protocole du jeu</span>
				</h2>
			</summary>
			<!-- Protocol block -->
			<div class="blk">
				<div class="title">Entrée d'initialisation</div>
				<div class="text">
					<var>myId</var>: votre identifiant de joueur. 
					<const>0</const>
					ou 
					<const>1</const>
					.<br/>
					<var>width</var>: nombre de cases dans une ligne de la carte.<br>
					<var>height</var>: nombre de cases dans une colonne de la carte.<br>
					<span class="statement-lineno">Prochaines <var>height</var>*<var>width</var> lignes :</span> deux entiers pour
					décrire chaque case de la carte, de gauche à droite, de haut en bas :
					<ul style="margin-top: 0">
						<li><var>regionId</var>: l’identifiant de la région dont fait partie cette case.</li>
						<li>
							<var>type</var>: le type de terrain de cette case (
							<const>0-2</const>
							)
						</li>
					</ul>
					<var>townCount</var>: nombre de villes sur la carte.<br>
					<span class="statement-lineno">Prochaines <var>townCount</var> lignes :</span>
					<ul style="margin-top: 0;">
						<li><var>townId</var>: identifiant unique de cette ville.</li>
						<li>
							<var>townX</var>: position X de cette ville (
							<const>0</const>
							est la plus à gauche).
						</li>
						<li>
							<var>townY</var>: position Y de cette ville (
							<const>0</const>
							est la plus en haut).
						</li>
						<li>
							<var>desiredConnections</var>:
							<ul style="margin-top: 0;">
								<li style="list-style-type: circle">
									Une chaîne de <var>townIds</var> séparés par des virgules.
									ex. "
									<const>1,2,4</const>
									"
								</li>
								<li style="list-style-type: circle">
									"
									<const>x</const>
									" si cette ville n’a pas de connexions souhaitées.
								</li>
							</ul>
						</li>
					</ul>
				</div>
			</div>
			<div class="blk">
				<div class="title">Entrée pour un tour de jeu</div>
				<div class="text">
					<var>myScore</var>: vos points.<br/>
					<var>foeScore</var>: points de votre adversaire.<br/>
					<span class="statement-lineno">Prochaines <var>height</var>*<var>width</var> lignes :</span>
					état de chaque case, dans le même ordre que précédemment :
					<ul style="margin-top:0">
						<li>
							<var>trackOwner</var>:
							<ul style="margin-top:0">
								<li style="list-style-type: circle">
									<const>-1</const>
									si cette case n’a pas de rail.
								</li>
								<li style="list-style-type: circle">
									<const>0</const>
									si le joueur 
									<const>0</const>
									possède un rail sur cette case.
								</li>
								<li style="list-style-type: circle">
									<const>1</const>
									si le joueur 
									<const>1</const>
									possède un rail sur cette case.
								</li>
								<li style="list-style-type: circle">
									<const>2</const>
									si un rail neutre est sur cette case.
								</li>
							</ul>
							<!-- BEGIN level1 -->
						<li><var>instability</var>: <em>Non utilisé dans cette ligue</em>.</li>
						<li><var>inked</var>: <em>Non utilisé dans cette ligue</em>. </li>
						<!-- END -->
						<!-- BEGIN level2 level3 -->
						<li><var>instability</var>: instabilité de la région dont fait partie cette case.
						</li>
						<li>
							<var>inked</var>: 
							<const>1</const>
							si la région a été effacée à l’encre (instabilité
							&ge;
							<const>4</const>
							), 
							<const>0</const>
							sinon. 
						</li>
						<!-- END -->
						<li>
							<!-- BEGIN level1 -->
							<em>(Pas utile dans cette ligue) </em>
							<!-- END -->
							<var>partOfActiveConnections</var>:
							<ul>
								<li style="list-style-type: circle">
									Une chaîne de paires <var>townId</var> séparées par des virgules indiquant que cette case fait partie
									d’une connexion active entre ces deux villes. <br>ex. "
									<const>
										1-2,1-3,4-7
									</const>
									":
									<em>la case fait partie du chemin le plus court entre les villes 1 & 2, villes 1 & 3, et villes 4 & 7.</em>
								</li>
								<li style="list-style-type: circle">
									"
									<const>x</const>
									" si cette case ne fait partie d’aucune connexion active.
								</li>
							</ul>
						</li>
					</ul>
				</div>
			</div>
			<!-- Protocol block -->
			<div class="blk">
				<div class="title">Sortie</div>
				<div class="text">
					Une seule ligne contenant au moins une action et au maximum une seule action 
					<action>AUTOPLACE</action>.<br/>
					Toutes les actions doivent être séparées par un point-virgule 
					<action>;</action>
					et être l’une des suivantes :
					<ul style="margin-bottom: 0">
						<li>
							<action>PLACE_TRACKS</action>
							suivie des coordonnées de l’emplacement désiré.
						</li>
						<li>
							<action>AUTOPLACE</action>
							suivie de deux paires de coordonnées, pour créer le chemin le moins coûteux entre les deux.
						</li>
					</ul>
					<!-- BEGIN level2 level3 -->
					<!-- BEGIN level2 -->
					<div class="statement-new-league-rule">
						<!-- END -->
						<ul style="margin-top: 0; margin-bottom: 0">
							<li>
								<action>DISRUPT</action>
								suivie du <var>regionId</var> de la région à perturber. <em>Remarque :</em> remplacez <var>regionId</var> par les coordonnées <var>x</var>,<var>y</var> pour perturber la région à cet emplacement.
							</li>
						</ul>
						<!-- BEGIN level2 -->
					</div>
					<!-- END -->
					<!-- END -->
					<ul style="margin-top: 0; margin-bottom: 0;">
						<li>
							<action>MESSAGE</action>
							suivie d’un texte, à afficher dans le visualiseur.
						</li>
					</ul>
					<ul style="margin-top: 0">
						<li>
							<action>WAIT</action>
						</li>
					</ul>
				</div>
			</div>
			<div class="blk">
				<div class="title">Contraintes</div>
				<div class="text">
					Temps de réponse par tour ≤ 
					<const>50</const>
					ms <br />
					Temps de réponse pour le premier tour ≤ 
					<const>1000</const>
					ms
					<br />
					<const>21</const>
					&le; <var>width</var> &le; 
					<const>30</const>
					<br />
					<const>14</const>
					&le; <var>height</var> &le; 
					<const>20</const>
					<br />
					<const>4</const>
					&le; <var>townCount</var> &le; 
					<const>12</const>
				</div>
			</div>
			<!-- BEGIN level1 level2 -->
		</details>
		<!-- END -->
		<!-- BEGIN level3 -->
	</details>

	<!-- END -->
	    <!-- STORY -->
    <div class="statement-story-background">
        <div class="statement-story-cover"
            style="background-size: cover; background-image: url(/servlet/fileservlet?id=150393782203046)">
            <div class="statement-story" style="min-height: 300px; position: relative">
                <h2><span style="color: #b3b9ad">Pour Démarrer</span></h2>
                <div class="story-text">
                    Pourquoi ne pas se lancer dans la bataille avec l'un de ces <b>IA Starters</b>, fournis par
                    l'équipe&nbsp;:
      <ul>
        <li>C++
          <a style="color: #f2bb13; border-bottom: 1px dotted #f2bb13;"
          rel="nofollow" target="_blank" href="https://gist.github.com/CGjupoulton/bbd4b720ae7e0f1e5f2e970bb42ed066">https://gist.github.com/CGjupoulton/bbd4b720ae7e0f1e5f2e970bb42ed066</a>
        </li>
        
        <li>JavaScript
          <a style="color: #f2bb13; border-bottom: 1px dotted #f2bb13;"
          rel="nofollow" target="_blank" href="https://gist.github.com/CGjupoulton/279a0f6ce4995bf4c55391fd40ae8bff">https://gist.github.com/CGjupoulton/279a0f6ce4995bf4c55391fd40ae8bff</a>
        </li>
        <li>Java
          <a style="color: #f2bb13; border-bottom: 1px dotted #f2bb13;"
          rel="nofollow" target="_blank" href="https://gist.github.com/CGjupoulton/17397683833e324a7b8c7c7c642239d3">https://gist.github.com/CGjupoulton/17397683833e324a7b8c7c7c642239d3</a>
        </li>
        <li>Python
          <a style="color: #f2bb13; border-bottom: 1px dotted #f2bb13;"
          rel="nofollow" target="_blank" href="https://gist.github.com/CGjupoulton/5a73bbd1142af98c6ca6887648b07cc2">https://gist.github.com/CGjupoulton/5a73bbd1142af98c6ca6887648b07cc2</a>
        </li>
        <!-- <li>TypeScript
          <a style="color: #f2bb13; border-bottom: 1px dotted #f2bb13;"
          rel="nofollow" target="_blank" href="tba">tba</a>
        </li> -->
      </ul>
                    <p>
                        Vous pouvez les modifier selon votre style, ou les prendre comme exemple pour tout coder à
                        partir de
                        zero.
                    </p>
                </div>
            </div>
        </div>
    </div>
	<!-- SHOW_SAVE_PDF_BUTTON -->
</div>