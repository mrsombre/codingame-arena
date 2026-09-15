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
			This is a <b>league based</b> challenge.
		</p>
		<span class="statement-league-alert-content">
		For this challenge, multiple leagues for the same game are available. Once
		you have proven your skills against the first Boss, you will access a
		higher league and extra rules will be available.
		<br /><br />
		In the first few leagues, your submission will only fight the boss in
		the arena. Win a best-of-five to advance.
		</span>
	</div>
	<!-- GOAL -->
	<div class="statement-section statement-goal">
		<h2>
			<span class="icon icon-goal">&nbsp;</span>
			<span>Goal</span>
		</h2>
		<div class="statement-goal-content">
			<div>
				<!-- BEGIN level1 -->
				The first two Leagues have a special <b>objective</b> to achieve. The full game will play, but
				you can only win by completing the objective. Once you do, you can start work on your complete bot.
				<!-- END -->
				<!-- BEGIN level3 -->
				<p>Score more points than your opponent by drawing railway lines between towns while sabotaging your
					opponent's efforts.
				</p>
				<!-- END -->
			
				<!-- BEGIN level1 -->
				<h3 style="
					font-size: 16px;
					font-weight: 700;
					padding-top: 20px;
					color: #838891;
					padding-bottom: 15px;
					">
					🎯 League Objective 1:
				</h3>
				Connect two <b>towns</b> to form an <b>active train connection</b> to instantly win the game.<br><br>
				The <b>Boss</b> AI will skip their turns. If you fail to form a connection within 
				<const>100</const>
				turns, you will
				lose. Win at least 
				<const>3</const>
				or more times out of 
				<const>5</const>
				to progress to the next
				league.
				<!-- END -->
				<!-- BEGIN level2 -->
				<h3 style="
					font-size: 16px;
					font-weight: 700;
					padding-top: 20px;
					color: #838891;
					padding-bottom: 15px;
					">
					🎯 League Objective 2:
				</h3>
				Use the 
				<action>DISRUPT</action>
				action to wash away at least one of your opponent's tracks.<br>
				<br>
				The <b>Boss</b> AI will place tracks between random towns. To win, disrupt any region containing an enemy track enough
				to cause it to get <b>inked out</b>, destroying any tracks within. Win at least 
				<const>3</const>
				or
				more times out of 
				<const>5</const>
				to progress to the next
				league. <em>More information on disruption lower down.</em>
				<!-- END -->
			</div>
		</div>
	</div>
	<!-- RULES -->
	<div class="statement-section statement-rules">
		<h2>
			<span class="icon icon-rules">&nbsp;</span>
			<span>Rules</span>
		</h2>
		<div class="statement-rules-content">
			<p>In this game both players use <b>paint</b> to draw train tracks on a <b>magic</b> map, connecting
				towns on the map will bring prosperity to your own world.
			</p>
			<p>The map is represented in the game by a <b>grid</b>.</p>
			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🗺️ Map
			</h3>
			<p>
				The grid is made up of cells that can have one of three types:
			<ul>
				<li>
					Type 
					<const>0</const>
					for <b>plains</b>.
				</li>
				<li>
					Type 
					<const>1</const>
					for <b>river</b>.
				</li>
				<li>
					Type 
					<const>2</const>
					for <b>mountains</b>.
				</li>
			</ul>
			</p>
			<p>
				The grid is partitioned into <b>regions</b>. Each region is made up of multiple contiguous
				cells. Each region has a unique <var>regionId</var>. Regions are susceptible to
				<b>disruption</b> by players.
				<!-- BEGIN level1 -->
				<em>
				More information on disruption in next league.
				</em>
				<!-- END -->
				<!-- BEGIN level2 level3 -->
				<br><br/><em>More information on disruption lower down.</em>
				<!-- END -->
			</p>
			<p>
				Some regions will contain a <b>town</b>. Towns can only be found on <b>plain</b> cells.
			</p>
			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🏯 Towns
			</h3>
			<p>
				Each game starts with multiple <b>towns</b> placed randomly across the map. There will only be one
				per <b>region</b> and no two regions sharing a border will both contain a town.
			</p>
			<p>
				Each town has a unique <var>townId</var>.
			</p>
			<p>
				Each town will have a list of <var>desiredConnections</var>: a list of town ids representing all the
				<b>other towns</b> this town would like to be <b>connected to</b> via <b>train tracks</b> placed by
				players.
			</p>
			<p>Providing a town with <b>train tracks</b> connecting it to a desired town is how players score
				<b>points</b>.
			</p>
			<p>
				Desired connections are <b>unilateral</b>. Meaning if town 
				<const>0</const>
				spawns with a desired
				connection to town 
				<const>1</const>
				, town 
				<const>1</const>
				<b>will not</b> want to connect to town
				<const>0</const>
				.
			</p>
			<p>
				A town can have 
				<const>zero</const>
				<var>desiredConnections</var>, but will always be the subject of
				at least one other town's <var>desiredConnections</var>.
			</p>
			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🛤️ Placing Train Tracks
			</h3>
			<p>
				Players are given on each turn 
				<const>3</const>
				<b>paint points</b> they can use to place
				<b>train tracks</b> on the map. ⚠️ These points do not carry over to the next turn and will be lost
				if left unused.
			</p>
			<p>
				It costs:
			<ul>
				<li>
					<const>1</const>
					paint point to place a track on <b>plains</b>.
				</li>
				<li>
					<const>2</const>
					paint points to place a track on <b>river</b>.
				</li>
				<li>
					<const>3</const>
					paint points to place a track on <b>mountains</b>.
				</li>
			</ul>
			</p>
			<p>
				A track's <var>owner</var> is the <var>playerId</var> (
				<const>0</const>
				-
				<const>1</const>
				) of the
				player that placed it. They will be the same color.
			</p>
			<p>
				If both players place a track on the <b>same turn</b> at the <b>same location</b>, the track's
				<var>owner</var> will be 
				<const>2</const>
				, indicating a neutral track piece.
			</p>
			<p>A train track <b>cannot</b> be placed on a <b>town</b> or on an existing track.</p>
			<p>
				Once placed, a train track will <b>automatically connect</b> to other <b>tracks</b> and <b>towns</b>
				orthogonally adjacent to it.
			</p>
			<h3 style="
				font-size: 16px;
				font-weight: 700;
				padding-top: 20px;
				color: #838891;
				padding-bottom: 15px;
				">
				🏯🛤️🏯 Connections
			</h3>
			<p>
				For each pair of towns in which one has the other in its <var>desiredConnections</var>, if at least
				one <b>path</b> between the two exists, the shortest such <b>path</b> becomes the
				<b>active connection</b> between those towns.
			</p>
			<p> A <b>path</b> is an uninterrupted sequence of orthogonally adjacent cells with a <b>train track</b>
				or a
				<b>town</b>.
			</p>
			<p>If there are <b>multiple</b> shortest paths, the chosen path will always <b>prioritize</b> the
				direction in the following order when moving from the requesting town to the desired connected town:
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
				At the end of every turn, each <b>active connection</b> will provide 
				<const>1</const>
				point to each
				player for every track they <b>own</b> in the <b>path</b>.
			</p>
			<br>
			<p><strong>Example 1:</strong></p>
			<div style="text-align: center; margin: 15px">
				<img src="https://static.codingame.com/servlet/fileservlet?id=151670109832251"
					style="width: 60%; max-width: 400px;" />
			</div>
			<p> Here, there is an <b>active connection</b> from town <b>0</b> to town <b>1</b> and one from town
				<b>0</b> to town <b>2</b>.
			</p>
			<p>
				Both the red and the blue players will gain 
				<const>3</const>
				points for the <b>0-1</b>
				connection and 
				<const>4</const>
				points for the <b>0-2</b> connection at the end of the turn.
			</p>
			<br/>
			<p><strong>Example 2:</strong></p>
			<div style="text-align: center; margin: 15px">
				<img src="https://static.codingame.com/servlet/fileservlet?id=151670123783140"
					style="width: 60%; max-width: 400px" />
			</div>
			<p>
				Here, only the shortest path from <b>0</b> to <b>2</b> is used for connection <b>0-2</b>.
				Meaning the red player will now gain 
				<const>3</const>
				points for that connection, and the
				blue player will get none.
			</p>
			<p>
				There are two equally short paths from <b>0</b> to <b>1</b>, but since 
				<const>EAST</const>
				has a higher priority than 
				<const>SOUTH</const>
				, the chosen path will be the one that goes
				through town <b>2</b>. The red player will gain 
				<const>4</const>
				points for that connection,
				and the blue player will only gain 
				<const>1</const>
				.
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
					💥 Disruption
				</h3>
				<p>
					Simarly to <b>paint points</b>, players can also use 
					<const>1</const>
					<b>disruption</b>
					point
					per turn. It can be
					used to
					tamper with the map, giving you an edge over your opponent.
					<em>These points aren't retained in between turns either.</em>
				</p>
				<p>
					Players may spend their <b>disruption point</b> each turn to increase the
					<var>instability</var>
					of
					any <b>region</b> by 
					<const>1</const>
					.
				</p>
				<p>
					Once a region's <var>instability</var> reaches 
					<const>4</const>
					, that region is
					<b>inked out</b>, washing any placed <b>train tracks</b> away, and rendering any future
					placements
					on it <b>impossible</b>. Any active connections via this region will be severed.
				</p>
				<p>It is not possible to disrupt a region that is already <b>inked out</b>.</p>
				<p>Regions with a town cannot be disrupted.</p>
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
			<p>Each turn, players must provide at least one action on the standard output.</p>
			<p>
				Actions must be separated by a semicolon 
				<action>;</action>
				and be one of the following:
			</p>
			<ul style="margin-bottom: 0">
				<li>
					<action>PLACE_TRACKS x y</action>
					: place a track on a free cell.
				</li>
				<li>
					<action>AUTOPLACE fromX fromY toX toY</action>
					: automatically generates a list of
					actions
					for the <b>cheapest</b> path from <var>from</var> to <var>to</var> in terms of paint
					points.
					This will do nothing if a path already exists.
					<br/>
					<em>The generated actions replace this command.</em>
				</li>
			</ul>
			<!-- BEGIN level2 level3 -->
			<!-- BEGIN level2 -->
			<div class="statement-new-league-rule">
				<!-- END -->
				<ul style="margin-top: 0; margin-bottom: 0;">
					<li>
						<action>DISRUPT regionId</action>
						: increase the instability of a region.
						<em>Note:</em>
						<action>DISRUPT x y</action>
						also works, to target the region (x,y) is part of.
					</li>
				</ul>
				<!-- BEGIN level2 -->
			</div>
			<!-- END -->
			<!-- END -->
			<ul style="margin-top: 0">
				<li>
					<action>WAIT</action>
					: do nothing.
				</li>
			</ul>
			<br>
			<!-- Victory conditions -->
			<div class="statement-victory-conditions">
				<div class="icon victory"></div>
				<div class="blk">
					<div class="title">Victory Conditions</div>
					<div class="text">
						<ul>
							<!-- BEGIN level1 -->
							Form an <b>active connection</b> and score any number of points.
							<!-- END -->
							<!-- BEGIN level2 -->
							<li>
								Cause your opponent to lose at least one <b>train track</b> by causing a
								region
								to <b>ink out</b> using 
								<action>DISRUPT</action>
								.
							</li>
							<!-- END -->
							<!-- BEGIN level3 -->
							<li>
								Have the most points after 
								<const>100</const>
								turns.
							</li>
							<li>Be in the lead if all desired connections become impossible to fulfill.</li>
							<!-- END -->
						</ul>
					</div>
				</div>
			</div>
			<!-- Lose conditions -->
			<div class="statement-lose-conditions">
				<div class="icon lose"></div>
				<div class="blk">
					<div class="title">Defeat Conditions</div>
					<div class="text">
						<!-- BEGIN level1 level2 -->
						<ul>
							<li>Your program does not provide a command in the allotted time or one of the
								commands
								is invalid.
							</li>
							<li>
								You do not complete the objective within 
								<const>100</const>
								turns.
							</li>
						</ul>
						<!-- END -->
						<!-- BEGIN level3 -->
						Your program does not provide a command in the allotted time or one
						of the commands is invalid.
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
					<span>Technical Details</span>
				</h2>
				<div class="statement-expert-rules-content">
					<ul style="padding-left: 20px;padding-bottom: 0">
						<li>
							You can check out the source code of this game <a rel="nofollow" target="_blank"
							href="https://github.com/CGjupoulton/SummerChallenge2026">on this GitHub repo</a>.
						</li>
						<li>
							All 
							<action>PLACE_TRACKS</action>
							actions, including from 
							<action>AUTOPLACE
							</action>
							are performed <b>before</b>
							<action>DISRUPT</action>
							actions. Points are scored at the very end of a turn,
							<b>after</b> inking out unstable regions.
						</li>
						<li>
							Commands that are impossible actions are skipped. If an impossible action is
							part of
							an 
							<action>AUTOPLACE</action>
							, the rest of the generated actions are skipped
							even if
							they are possible.
						</li>
					</ul>
				</div>
			</div>
			<!-- END -->
			<div class="statement-section statement-expertrules">
				<h2>
					<span>🐞 Debugging tips</span>
				</h2>
				<ul>
					<li>
						Hover over the grid to see extra information on the cell under your
						mouse.
					</li>
					<!-- <li>
						Assign the special <action>MESSAGE text</action> action to an agent
						and that text will appear above your agent.
						</li> -->
					<li>
						Press the gear icon on the viewer to access extra display options.
					</li>
					<li>
						Use the keyboard to control the action: space to play/pause, arrows to
						step 1 frame at a time.
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
					>Click to expand</span>
				<h2 style="margin-bottom: 0">
					<span class="icon icon-protocol">&nbsp;</span>
					<span>Game Protocol</span>
				</h2>
			</summary>
			<!-- Protocol block -->
			<div class="blk">
				<div class="title">Initialization Input</div>
				<div class="text">
					<var>myId</var>: your player id. 
					<const>0</const>
					or 
					<const>1</const>
					.<br/>
					<var>width</var>: number of cells in a row of the map.<br>
					<var>height</var>: number of cells in a column of the map.<br>
					<span class="statement-lineno">Next <var>height</var>*<var>width</var> lines:</span> two
					integers to
					describe each cell of the map, from left to right, top to
					bottom:
					<ul style="margin-top: 0">
						<li><var>regionId</var>: the id of the region this cell is a part of.</li>
						<li>
							<var>type</var>: the terrain type of this cell (
							<const>0-2</const>
							)
						</li>
					</ul>
					<var>townCount</var>: number of towns on the map.<br>
					<span class="statement-lineno">Next <var>townCount</var> lines:</span>
					<ul style="margin-top: 0;">
						<li><var>townId</var>: unique identifier of this town.</li>
						<li>
							<var>townX</var>: X position of this town (
							<const>0</const>
							is left-most).
						</li>
						<li>
							<var>townY</var>: Y position of this town (
							<const>0</const>
							is top-most).
						</li>
						<li>
							<var>desiredConnections</var>:
							<ul style="margin-top: 0;">
								<li style="list-style-type: circle">
									A string of comma-separated
									<var>townIds</var>.
									e.g.
									"
									<const>1,2,4</const>
									"
								</li>
								<li style="list-style-type: circle">
									"
									<const>x</const>
									" if this town has no desired connections.
								</li>
							</ul>
						</li>
					</ul>
				</div>
			</div>
			<div class="blk">
				<div class="title">Input for one game turn</div>
				<div class="text">
					<var>myScore</var>: your points.<br/>
					<var>foeScore</var>: your opponent's points.<br/>
					<span class="statement-lineno">Next <var>height</var>*<var>width</var> lines:</span>
					state of
					each
					cell, in the same order they were given before:
					<ul style="margin-top:0">
						<li>
							<var>trackOwner</var>:
							<ul style="margin-top:0">
								<li style="list-style-type: circle">
									<const>-1</const>
									if this cell has no track.
								</li>
								<li style="list-style-type: circle">
									<const>0</const>
									if player 
									<const>0</const>
									owns a track on this cell.
								</li>
								<li style="list-style-type: circle">
									<const>1</const>
									if player 
									<const>1</const>
									owns a track on this cell.
								</li>
								<li style="list-style-type: circle">
									<const>2</const>
									if there is a neutral track on this cell.
								</li>
							</ul>
							<!-- BEGIN level1 -->
						<li><var>instability</var>: <em>Unused in this league</em>.</li>
						<li><var>inked</var>: <em>Unused in this league</em>. </li>
						<!-- END -->
						<!-- BEGIN level2 level3 -->
						<li><var>instability</var>: the instability of the region this cell is a part of.
						</li>
						<li>
							<var>inked</var>: 
							<const>1</const>
							if the region has been inked out (instability
							&ge;
							<const>4</const>
							), 
							<const>0</const>
							otherwise. 
						</li>
						<!-- END -->
						<li>
							<!-- BEGIN level1 -->
							<em>(Not useful in this league) </em>
							<!-- END -->
							<var>partOfActiveConnections</var>:
							<ul>
								<li style="list-style-type: circle">
									A string of comma-separated <var>townId</var> pairs indicating this cell
									is part
									of
									an active connection between those two towns. <br>e.g. "
									<const>
										1-2,1-3,4-7
									</const>
									":
									<em>cell is part of the shortest path between towns 1 & 2, towns 1 & 3, and towns 4 & 7.</em>
								</li>
								<li style="list-style-type: circle">
									"
									<const>x</const>
									" if this cell is not part of any active connection.
								</li>
							</ul>
						</li>
					</ul>
				</div>
			</div>
			<!-- Protocol block -->
			<div class="blk">
				<div class="title">Output</div>
				<div class="text">
					A single line containing at least one action and at most a single 
					<action>AUTOPLACE
					</action>
					action.<br/>
					All actions must be separated with a semicolon 
					<action>;</action>
					and be one of the following:
					<ul style="margin-bottom: 0">
						<li>
							<action>PLACE_TRACKS</action>
							followed by the coordinates of the desired
							location.
						</li>
						<li>
							<action>AUTOPLACE</action>
							followed by two pairs of coordinates, to create the
							cheapest
							path
							between the two.
						</li>
					</ul>
					<!-- BEGIN level2 level3 -->
					<!-- BEGIN level2 -->
					<div class="statement-new-league-rule">
						<!-- END -->
						<ul style="margin-top: 0; margin-bottom: 0">
							<li>
								<action>DISRUPT</action>
								followed by the <var>regionId</var> of the region
								you wish
								to
								disrupt. <em>Note:</em> replace <var>regionId</var> by <var>x</var>,<var>y</var> coordinates target the region at
								that location.
							</li>
						</ul>
						<!-- BEGIN level2 -->
					</div>
					<!-- END -->
					<!-- END -->
					<ul style="margin-top: 0; margin-bottom: 0;">
						<li>
							<action>MESSAGE</action>
							followed by text, to be displayed in the viewer.
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
				<div class="title">Constraints</div>
				<div class="text">
					Response time per turn ≤ 
					<const>50</const>
					ms <br />Response time for
					the first turn ≤ 
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
    <h2><span style="color: #b3b9ad">To Start</span></h2>
    <div class="story-text">
      Why not start the battle with one of these <b>AI Starters</b>, provided by the team:
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
        You can modify them to match your style or take them as an example to code everything from
        scratch.
      </p>
    </div>
  </div>
</div>
</div>
	<!-- SHOW_SAVE_PDF_BUTTON -->
</div>
