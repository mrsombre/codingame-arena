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
        Snag power sources to grow your snake-like robots and make sure you have the biggest bots standing.
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
      <span>Rules</span>
    </h2>

    <div class="statement-rules-content">
      <p>
        The game is played on a <b>grid</b>.
      </p>
      <p>
        Each player controls a team of <b>snakebots</b>. On each turn, the snakebots move <b>simultaneously</b> according to the players' commands.
      </p>


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
        The grid is seen from the side, and is made up of <b>platforms</b>. Platforms are impassable cells.
      </p>
      
      <p>
        On this grid you may find parts of a snakebot's <b>body</b> and <b>power sources</b>.
      
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
        Snakebots are multiple adjacent cell-sized body parts. The first cell being their <b>head</b>.
      </p>
      <p>
        Snakebots are affected by <b>gravity</b>. Meaning <b>at least one body</b> part must be above something <b>solid</b> or they will <b>fall</b>.
      </p>

      <p>
        Other snakebots are considered <b>solid</b>, as well as <b>platforms</b> and <b>power sources
      </b>.

            <h3 style="
        font-size: 16px;
        font-weight: 700;
        padding-top: 20px;
        color: #838891;
        padding-bottom: 15px;
        ">
        ↪️ Movement
      </h3>

      <p>
        Snakebots are <b>perpetually moving</b>, and will on each turn head in the last direction they were facing unless given a new direction to turn by the player. 
      </p>
      <p>The starting direction is <b>up</b>.

      <p>
        When moving, a snakebot will advance their head in the direction it is facing, and the rest of its body parts will follow.
      </p>

      <p>
        <b>Case 1:</b> If the cell the snakebot's head has moved into contains a <b>platform</b> or other <b>body part</b>, the snakebot's head is destroyed and the next part in its body becomes the <b>new head</b>. This only happens if is has at least three remaining parts. If not, the whole snakebot is removed.
      </p>

      <p>
        <b>Case 2:</b> If the cell the snakebot has moved into contains a <b>power source</b>, the snakebot will <b>eat it</b>. This has two effects:
        <ol>
          <li>The snakebot <b>grows</b>, a new body part appears at the end of its body.</li>
          <li>This cell is no longer considered <b>solid</b>.</li>
        </ol>

    </p>
  <p>
    These collisions are resolved simultaneously for all snakebots.
    </p>
    <p>
      <b>Special case:</b> If multiple snakebot heads collide on a cell containing a <b>power source</b>, that power source is considered <b>eaten by each of those snakebots</b>!
    </p>

      <p>
        Once movement and removals are resolved, the snakebots all <b>fall</b> downwards until a body part lands on something <b>solid</b>.
      </p>

      <p>
        It is possible to extend your <b>snakebot</b> beyond the borders of the grid. But falling out of the playing area will result in the snakebot being removed.
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
      <p>Each turn, players must provide at least one action on the standard output.</p>
      <p>
        Actions must be separated by a semicolon 
        <action>;</action>
        and be one of the following.
      </p>

      Commands to move snakebot with id <var>id</var>:

      <ul style="margin-bottom: 0">
        <li>
          <action>id UP</action>: sets a snakebot's direction to UP <const>(x,y)=(0,-1)</const>.
        </li>
        <li>
          <action>id DOWN</action>: sets a snakebot's direction to DOWN <const>(x,y)=(0,1)</const>.
        </li>
        <li>
          <action>id LEFT</action>: sets a snakebot's direction to LEFT <const>(x,y)=(-1,0)</const>.
        </li>
        <li>
          <action>id RIGHT</action>: sets a snakebot's direction to RIGHT <const>(x,y)=(1,0)</const>.
        </li>
      </ul>
      <p>
      Any of the movement actions can be followed by text which will be displayed above the appropriate snakebot for <b>debugging</b> purposes.
      </p>
      <p>
        Special commands:
      </p>
      <ul style="margin-bottom: 0">
        <li>
          <action>MARK x y</action>: places a marker at the specified coordinates. Markers are visible in the viewer for <b>debugging</b> purposes.
        </li>
        <li>
          <action>WAIT</action>
          : do nothing.
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
        ⛔ Game End
      </h3>
      The game is over if any of these are true at the end of a turn:
      <ul style="margin-bottom: 0">
        <li>All of a player's snakebots have been removed.</li>
        <li>There are no more <b>power sources</b> left to eat.</li>
        <li>200 turns have passed.</li>
      </ul>

      <br>
      <!-- Victory conditions -->
      <div class="statement-victory-conditions">
        <div class="icon victory"></div>
        <div class="blk">
          <div class="title">Victory Conditions</div>
          <div class="text">
              Have more total body parts across all your snakebots than your opponent at the end of the game.
          </div>
        </div>
      </div>
      <!-- Lose conditions -->
      <div class="statement-lose-conditions">
        <div class="icon lose"></div>
        <div class="blk">
          <div class="title">Defeat Conditions</div>
          <div class="text">
            Your program does not provide a command in the alloted time or one
            of the commands is invalid.
          </div>
        </div>
      </div>
      <br />

      <!-- EXPERT RULES -->
      <div class="statement-section statement-expertrules">
        <h2>
          <span class="icon icon-expertrules">&nbsp;</span>
          <span>Technical Details</span>
        </h2>
        <div class="statement-expert-rules-content">
          <ul style="padding-left: 20px;padding-bottom: 0">
            <li>
              The source code of this game will be published to <a rel="nofollow" target="_blank"
              href="https://github.com/CodinGame/WinterChallenge2026-Exotec">this GitHub repo</a>.
              </li>
            
          </ul>
        </div>
      </div>

      <div class="statement-section statement-expertrules">
        <h2>
          <span>🐞 Debugging tips</span>
        </h2>
        <ul>
          <li>
            Use the <action>MARK x y</action> action to highlight up to 4 cells per turn.
          </li>
          <li>
            Hover over the grid to see extra information on the cell under your
            mouse.
          </li>

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
  <details class="statement-section statement-protocol">
    
      <summary
        open
        style="cursor: pointer; margin-bottom: 10px; display: inline-block"
      >
        <span style="display: inline-block; margin-bottom: 10px"
          >Click to expand</span
        >
        <h2 style="margin-bottom: 0">
          <span class="icon icon-protocol">&nbsp;</span>
          <span>Game Protocol</span>
        </h2>
      </summary>

      <!-- Protocol block -->
      <div class="blk">
        <div class="title">Initialization Input</div>
        <div class="text">
          <span class="statement-lineno">Line 1:</span> <var>myId</var>, an integer
 for your player identification (<const>0</const> or <const>1</const>).
            <br/>
            <span class="statement-lineno">Line 2:</span> <var>width</var>, an integer for the width of the grid.
          <br/>
            <span class="statement-lineno">Line 3:</span> <var>height</var> an integer for the height of the grid.

          <br/>
          <br/>
            <span class="statement-lineno">Next </span><var>height</var>
            
            <span class="statement-lineno">lines:</span> One row of the grid 
            containing
            <var>width</var> characters each. Each character can be:
          
          <ul>
            <li><const>#</const>: a <b>platform</b>.</li>
            <li><const>.</const>: a <b>free</b> cell.</li>
          </ul>
          
          <span class="statement-lineno">Next line:</span> <var>snakbotsPerPlayer</var> an integer for amount of snakebots each player controls.
          <br/>
          <br/>
          <span class="statement-lineno">Next </span><var>snakbotsPerPlayer</var>
          
          <span class="statement-lineno">lines:</span> One integer <var>snakebotId</var> for each of the snakebots controlled by <b>you</b>.
          
          <br/><br/>
          <span class="statement-lineno">Next </span><var>snakbotsPerPlayer</var>
          <span class="statement-lineno">lines:</span> One integer <var>snakebotId</var> for each of the snakebots controlled by <b>your opponent</b>.
        </div>
      </div>
      <div class="blk">
        <div class="title">Input for one game turn</div>
        <div class="text">
          <p>
            <span class="statement-lineno">First line:</span> <var>powerSourceCount</var>, one integer
             for the number of remaining <b>power sources</b> on the
            grid.
          </p>
          <p>
            <span class="statement-lineno">Next </span><var>powerSourceCount</var>
            <span class="statement-lineno">lines:</span> The following
            <const>2</const> integers for each power source:
          </p>
          <ul>
            <li><var>x</var>: X coordinate</li>
            <li><var>y</var>: Y coordinate</li>
          </ul>
          <p>
            <span class="statement-lineno">Next line:</span> <var>snakebotCount</var>, one integer
             for the number of remaining <b>snakebots</b> on the
            grid.
          </p>
          <p>
            <span class="statement-lineno">Next </span><var>snakebotCount</var>
            <span class="statement-lineno">lines:</span> The following
            <const>2</const> inputs for each snakebot:
          </p>
          <ul>
            <li><var>snakebotId</var>: an integer for this snakebot's identifier</li>
            <li><var>body</var>: a string of colon (<const>:</const>) separated coordinates, each coordinate being the position of a body part with the form "<const>x,y</const>". The first coordinate in the list is the <b>head</b>.<br>
            <em>Example: </em>"<const>0,1:1,1:2,1</const>"<em> is a snakebot with a head at <const>x=0</const> and <const>y=1</const> with two more body parts to the right.</em></li>
          </ul>          
        </div>
      </div>

      <!-- Protocol block -->
      <div class="blk">
        <div class="title">Output</div>
        <div class="text">
          <p>
            A single line, containing at least one action:
          </p>
          <!-- <em>(At least on action, and no more than thirty).</em> -->
          <ul>
            <li>
              <action>id UP</action> where id is the <var>snakebotId</var> of a snakebot you control.
            </li>
            <li>
              <action>id DOWN</action> where id is the <var>snakebotId</var> of a snakebot you control.
            </li>
            <li>
              <action>id LEFT</action> where id is the <var>snakebotId</var> of a snakebot you control.
            </li>
            <li>
              <action>id RIGHT</action> where id is the <var>snakebotId</var> of a snakebot you control.
            </li>
            <li>
              <action>MARK x y</action>: you may mark up to <b>four</b> coordinates per turn.
            </li>
            <li>
              <action>WAIT</action>
            </li>
          </ul>
          <p>
            Instructions are separated by semi-columns (<const>;</const>). For example:
          </p>
          <p>
            <action><code>1 LEFT;2 RIGHT;MARK 12 2</code></action>
          </p>
        </div>
      </div>

      <div class="blk">
        <div class="title">Constraints</div>
        <div class="text">
          Response time per turn ≤ <const>50</const>ms <br />Response time for
          the first turn ≤ <const>1000</const>ms
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
