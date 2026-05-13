<!-- LEAGUES level1 level2 level3 level4 -->
<div id="statement_back" class="statement_back" style="display:none" xmlns="http://www.w3.org/1999/html"></div>
<div class="statement-body">
  <!-- BEGIN level1 level2 level3 -->
  <div style="color: #7cc576;
        background-color: rgba(124, 197, 118,.1);
        padding: 20px;
        margin-right: 15px;
        margin-left: 15px;
        margin-bottom: 10px;
        text-align: left;">
    <div style="text-align: center; margin-bottom: 6px">
      <img src="//cdn.codingame.com/smash-the-code/statement/league_wood_04.png" alt="Wood League"/>
    </div>
    <p style="text-align: center; font-weight: 700; margin-bottom: 6px;">
      <!-- BEGIN level1 -->
      This is a <b>league-based</b> challenge.
      <!-- END -->
      <!-- BEGIN level2 -->
      Welcome to the Wood1 league!
      <!-- END -->
      <!-- BEGIN level3 -->
      Welcome to the Bronze league!
      <!-- END -->
    </p>
    <span class="statement-league-alert-content">
      <!-- BEGIN level1 -->
      Wood leagues should be considered as a tutorial which lets players discover the different rules of the game. <br>
      In Bronze league all rules will be unlocked and the real challenge will begin.
      <!-- END -->
      <!-- BEGIN level2 -->
      In Wood 1 you can now <action>PLANT</action> trees and <action>TRAIN</action> new trolls.
      <!-- END -->
      <!-- BEGIN level3 -->
      In Bronze you can now <action>CHOP</action> trees for wood and <action>MINE</action> iron. Trees grow faster near water. Games last <const>300</const> turns. <br>
      There will be no additional rules in higher leagues.
      <!-- END -->
    </span>
    <!-- BEGIN level1 level2 -->
    <br>To see the full rules, go <a href="https://www.codingame.com/demo/1423805f9c28c9654cfd4d57b0208df71315c49">here</a> and switch leagues at the top-right corner.
    <!-- END -->
  </div>
  <!-- END -->

  <!-- BEGIN level3 -->
  <div style="color: #7c76c5;
      background-color: rgba(124, 118, 197, .1);
      padding: 20px;
      margin-right: 15px;
      margin-left: 15px;
      margin-bottom: 10px;
      text-align: left;">
      <p style="text-align: center; font-weight: 700; margin-bottom: 6px;">
        Balancing update: Maximum distance between shacks got increased. <br>
        For details see <a href="https://github.com/eulerscheZahl/Troll-Farm/commits/master/">github</a>, a <a href="https://github.com/eulerscheZahl/Troll-Farm/releases">new release</a> for local testing is available there as well.
      </p>
  </div>
  <!-- END -->

  <!-- GOAL -->
  <div class="statement-section statement-goal">
    <h1>
      <span class="icon icon-goal">&nbsp;</span>
      <span>The Goal</span>
    </h1>
    <div class="statement-goal-content">
      You control a pack of trolls. Make them collect the most resources.
    </div>
  </div>
  <!-- RULES -->
  <div class="statement-section statement-rules">
    <h1>
      <span class="icon icon-rules">&nbsp;</span>
      <span>Rules</span>
    </h1>
    <div>
      <div class="statement-rules-content">
        Each player starts with a shack and a troll. Control your troll to collect fruits
        <!-- BEGIN level3 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
                , mine iron and chop trees
        </span>
        <!-- END -->
        <!-- BEGIN level4 -->
        , mine iron and chop trees
        <!-- END -->
        and then bring the resources back to your shack. The resources can be used to train more trolls or plant trees. The game is played on a grid, cells have different terrains and trees on them.

        <br><br>
        <strong>Troll attributes</strong><br>
        Each troll has <const>4</const> attributes: <var>movementSpeed</var>, <var>carryCapacity</var>, <var>harvestPower</var> and <var>chopPower</var>. They can't be changed.

        <br><br>
        <strong>Trees</strong><br>
        A tree has a <var>type</var> (<const>PLUM</const>, <const>LEMON</const>, <const>APPLE</const>, <const>BANANA</const>), a <var>size</var>, a <var>health</var>, an amount of <var>fruits</var> and a <var>cooldown</var>.
        When cooldown reaches <const>0</const>, the tree either grows in size or, if at maximum size (<const>4</const>), produces fruits. Trees can hold up to <const>3</const> fruits.
        <br><img src="https://cdn-games.codingame.com/community/1500515-1769333345408/26af7e3e9288d31081f823125c8e91471e07e63616a27fe4996d729d4f595ff1.png" style="width: 60%; max-width: 600px"/><br>
        <!-- BEGIN level3 level4 -->
        <!-- BEGIN level3 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        Trees grow faster when next to water. <br>
        <img src="https://cdn-games.codingame.com/community/1500515-1769333345408/f71ed0e8c9491ec980a15debac344034276fc7c3620f273165d548a4aba40e2e.png" style="width: 60%; max-width: 400px"/> <br>
        <br>
        Trees differ in cooldown and health as follows:
        <table>
          <tr><th></th><th style="min-width:70px">PLUM</th><th style="min-width:70px">LEMON</th><th style="min-width:70px">APPLE</th><th style="min-width:70px">BANANA</th></tr>
          <tr><td>Cooldown</td><td align="center"><const>8</const></td><td align="center"><const>8</const></td><td align="center"><const>9</const></td><td align="center"><const>6</const></td></tr>
          <tr><td>Cooldown near water</td><td align="center"><const>3</const></td><td align="center"><const>3</const></td><td align="center"><const>2</const></td><td align="center"><const>4</const></td></tr>
            <tr><td>Health (size 1)</td><td align="center"><const>6</const></td><td align="center"><const>6</const></td><td align="center"><const>11</const></td><td align="center"><const>3</const></td></tr>
            <tr><td>Health (size 2)</td><td align="center"><const>8</const></td><td align="center"><const>8</const></td><td align="center"><const>14</const></td><td align="center"><const>4</const></td></tr>
            <tr><td>Health (size 3)</td><td align="center"><const>10</const></td><td align="center"><const>10</const></td><td align="center"><const>17</const></td><td align="center"><const>5</const></td></tr>
            <tr><td>Health (size 4)</td><td align="center"><const>12</const></td><td align="center"><const>12</const></td><td align="center"><const>20</const></td><td align="center"><const>6</const></td></tr>
        </table>
        When a damaged tree grows, it gains the health difference between sizes.
        <!-- BEGIN level3 -->
        </span>
        <!-- END -->
        <!-- END -->
        

        <br><br>
        <strong>Moving units</strong><br>
        Each troll has a <var>movementSpeed</var> determining how many cells it can <action>MOVE</action> horizontally or vertically per turn. Each cell can hold at most <const>1</const> troll per team. Only <const>GRASS</const> cells are walkable.
        The shack is also the spawn point, but trolls can't walk back on that cell after leaving it.
        If the target is out of range or not walkable, the troll moves to the nearest reachable cell towards it.

        <br><br>
        <strong>Harvesting</strong><br>
        Trees can be harvested, given they have fruits on them. When a troll shares a cell with a tree, <action>HARVEST</action> will take as many fruits as possible by free <var>carryCapacity</var> and <var>harvestPower</var>.
        When two trolls harvest the same tree in parallel, they keep taking one fruit at a time, as long as there are fruits left and they can still harvest. The last fruit can get duplicated, so that both trolls can have it.

        <!-- BEGIN level2 level3 level4 -->
        <br><br>
        <!-- BEGIN level2 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        <strong>Planting trees</strong><br>
        A troll can <action>PLANT</action> a tree on its current cell if the troll carries at least <const>1</const> fruit of that type.
        If two trolls try to plant a tree of the same type on the same cell, both will lose a seed and the tree gets planted. If they try to plant different types, nothing will happen.
        <!-- BEGIN level2 -->
        </span>
        <!-- END -->
        <!-- END -->

        <!-- BEGIN level3 level4 -->
        <br><br>
        <!-- BEGIN level3 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        <strong>Chopping trees</strong><br>
        A troll can <action>CHOP</action> a tree on the same cell, reducing its health by <var>chopPower</var>. When health reaches <const>0</const>, the tree disappears and the troll collects wood equal to its size, capped by free <var>carryCapacity</var>.
        Any wood that the troll can't carry will just vanish. Similar to harvesting, both trolls get one wood at a time until they can't carry any more or the wood is completely split up. The last wood can get duplicated.
        <!-- BEGIN level3 -->
        </span>
        <!-- END -->
        <!-- END -->

        <!-- BEGIN level3 level4 -->
        <br><br>
        <!-- BEGIN level3 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        <strong>Mining</strong><br>
        A troll adjacent to an <const>IRON</const> cell can <action>MINE</action>, gaining up to <var>chopPower</var> iron limited by free <var>carryCapacity</var>. Iron resources are infinite.
        <!-- BEGIN level3 -->
        </span>
        <!-- END -->
        <!-- END -->

        <br><br>
        <strong>Dropping items</strong><br>
        When a troll carrying resources is next to its shack (horizontally or vertically), <action>DROP</action> transfers all carried items to the shack.

        <!-- BEGIN level2 level3 level4 -->
        <br><br>
        <!-- BEGIN level2 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        <strong>Picking items</strong><br>
        A troll can <action>PICK</action> one item from its shack. Only <const>1</const> item can be picked at a time.
        <!-- BEGIN level2 -->
        </span>
        <!-- END -->
        <!-- END -->

        <!-- BEGIN level2 level3 level4 -->
        <br><br>
        <!-- BEGIN level2 -->
        <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
        <!-- END -->
        <strong>Training trolls</strong><br>
        Use <action>TRAIN</action> to spawn a new troll at your shack and set its four attributes. Each attribute requires a different resource:
        <ul>
          <li><const>PLUM</const> for <var>movementSpeed</var></li>
          <li><const>LEMON</const> for <var>carryCapacity</var></li>
          <li><const>APPLE</const> for <var>harvestPower</var></li>
          <!-- BEGIN level2 -->
          <li>(reserved for higher leagues)</li>
          <!-- END -->
          <!-- BEGIN level3 level4 -->
          <li>
          <!-- BEGIN level3 -->
          <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
          <!-- END -->
          <const>IRON</const> for <var>chopPower</var>
          <!-- BEGIN level3 -->
          </span>
          <!-- END -->
          </li>
          <!-- END -->
        </ul>
        The cost for each attribute is equal to the number of existing trolls in the own team plus the attribute squared. <br>
        Example: you already have <b>2</b> trolls and want to <action>TRAIN</action> the following, giving you these costs: <br>
        <ul>
            <li> <b>2</b> <var>movementSpeed</var>: <b>2 + 2×2 = 6</b> <const>PLUM</const>s </li>
            <li> <b>3</b> <var>carryCapacity</var>: <b>2 + 3×3 = 11</b> <const>LEMON</const>s </li>
            <li> <b>1</b> <var>harvestPower</var>: <b>2 + 1×1 = 3</b> <const>APPLE</const>s </li>
          <!-- BEGIN level3 level4 -->
          <li>
          <!-- BEGIN level3 -->
          <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
          <!-- END -->
          <b>0</b> <var>chopPower</var>: <b>2 + 0×0 = 2</b> <const>IRON</const>s
          <!-- BEGIN level3 -->
          </span>
          <!-- END -->
          </li>
          <!-- END -->
        </ul>
        <!-- BEGIN level2 -->
        </span>
        <!-- END -->
        <!-- END -->

        <br><br>
      </div>
    </div>
  </div>

  <!-- Victory conditions -->
  <div class="statement-victory-conditions">
    <div class="icon victory"></div>
    <div class="blk">
      <div class="title">Victory Conditions</div>
      <div class="text">
        <ul style="padding-bottom: 0;">
          <li>
            You score more points than your opponent. Each fruit in your shack scores <const>1</const> point
            <!-- BEGIN level3 level4 -->
            ; <const>WOOD</const> scores <const>4</const> points. <const>IRON</const> has no score value.
            <!-- END -->
          </li>
        </ul>
      </div>
    </div>
  </div>
  <!-- Lose conditions -->
  <div class="statement-lose-conditions">
    <div class="icon lose"></div>
    <div class="blk">
      <div class="title">Loss Conditions</div>
      <div class="text">
        <ul style="padding-bottom: 0;">
          <li>
            You have fewer points than your opponent.
          </li>
          <li>
            You fail to respond in time or output an unrecognized command.
          </li>
        </ul>
      </div>
    </div>
  </div>
  <br>

  <!-- BEGIN level3 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
  <!-- BEGIN level3 level4 -->
  <!-- EXPERT RULES -->
  <div class="statement-section statement-expertrules">
    <h1>
      <span class="icon icon-expertrules">&nbsp;</span>
      <span>Expert Rules</span>
    </h1>
    <div class="statement-expert-rules-content">
      Referee source code: <a href="https://github.com/eulerscheZahl/Troll-Farm">https://github.com/eulerscheZahl/Troll-Farm</a> <br>
      Don't hesitate to change the viewer's options to help debug your code (<img style="opacity:.8;background:#20252a;" height="18" src="https://www.codingame.com/servlet/fileservlet?id=3463235186409" width="18">).
      <br> <br>
      The game turn order is:
      <ol>
        <li>Move trolls</li>
        <li>Harvest</li>
        <li>Plant</li>
        <li>Chop</li>
        <li>Pick</li>
        <li>Train</li>
        <li>Drop</li>
        <li>Mine</li>
        <li>Grow trees</li>
      </ol>
      Actions of the same type occur simultaneously. If both teams harvest or chop the same tree at once, resources are shared equally.<br>
      Then both get one item at a time until they can't carry any more or the resource is exhausted. The last item can get duplicated. <br>
      Each troll can only perform one action per turn. <br>
    </div>
  </div>
  <!-- END -->
  <!-- BEGIN level3 -->
  </span>
  <!-- END -->

  <!-- PROTOCOL -->
  <div class="statement-section statement-protocol">
    <h2>
      <span class="icon icon-protocol">&nbsp;</span>
      <span>Game Input</span>
    </h2>
    <!-- Protocol block -->
    <div class="blk">
      <div class="title">Initial input</div>
      <div class="text">

        <span class="statement-lineno">First line</span>: <var>width</var> <var>height</var>, the grid size.
        <br>
        <span class="statement-lineno">Next <var>height</var> lines</span>: each line has <var>width</var> characters:
<!-- BEGIN level1 level2 -->
In this league every character is <const>.</const> for <const>GRASS</const>, <const>0</const> for your own <const>SHACK</const>, <const>1</const> for your opponent's <const>SHACK</const>.
<!-- END -->
<!-- BEGIN level3 level4 -->
  <!-- BEGIN level3 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
           <const>.</const> for <const>GRASS</const>, <const>~</const> for <const>WATER</const>, <const>#</const> for <const>ROCK</const>, <const>+</const> for <const>IRON</const>, <const>0</const> for your own <const>SHACK</const>, <const>1</const> for your opponent's <const>SHACK</const>.
  <!-- BEGIN level3 -->
  </span>
  <!-- END -->
  <!-- END -->
      </div>

      <br>
      <br>
      <div class="title">Input for each game turn</div>
      <div class="text">
        <span class="statement-lineno">First two lines: your inventory followed by your opponent's inventory</span>: <var>plums</var>, <var>lemons</var>, <var>apples</var>, <var>bananas</var>,
  <!-- BEGIN level1 level2 -->
  <var>(reserved)</var>, <var>(reserved)</var>. In this league the last two values are always <const>0</const>.
  <!-- END -->
  <!-- BEGIN level3 level4 -->
  <!-- BEGIN level3 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
        <var>iron</var>, <var>wood</var>
  <!-- BEGIN level3 -->
  </span>
  <!-- END -->
  <!-- END -->

        <br>
        <span class="statement-lineno">Next line</span>: <var>treeCount</var>, the number of trees.
        <br>
        <span class="statement-lineno">Next <var>treeCount</var> lines</span>: <var>type</var> <var>x</var> <var>y</var> <var>size</var> <var>health</var> <var>fruits</var> <var>cooldown</var>, where <var>type</var> can be <const>PLUM</const>, <const>LEMON</const>, <const>APPLE</const>, <const>BANANA</const>.
        <br>
        <span class="statement-lineno">Next line</span>: <var>trollsCount</var>, the number of trolls.
        <br>
        <span class="statement-lineno">Next <var>trollsCount</var> lines</span>: <var>id</var>
        <var>player</var> <var>x</var> <var>y</var> <var>movementSpeed</var> <var>carryCapacity</var> <var>harvestPower</var>
  <!-- BEGIN level1 level2 -->
  <var>(reserved)</var>
  <!-- END -->
  <!-- BEGIN level3 level4 -->
        <var>chopPower</var>
  <!-- END -->
        <var>carryPlum</var> <var>carryLemon</var> <var>carryApple</var> <var>carryBanana</var>
  <!-- BEGIN level1 level2 -->
  <var>(reserved)</var> <var>(reserved)</var>,
  <!-- END -->
  <!-- BEGIN level3 level4 -->
        <var>carryIron</var> <var>carryWood</var>,
  <!-- END -->
        where <var>player</var> is <const>0</const> if you own the troll, <const>1</const> otherwise.
        <br>
        <br>
      </div>

      <div class="title">Output for one game turn</div>
      <div class="text">
        Each turn you can print any number of commands, separated by
        <action>;</action>.
        <ul style="padding-bottom: 0;">
          <li>
            <action>MOVE id x y</action>
            Move troll <var>id</var> to cell (<var>x</var>, <var>y</var>).
          </li>

          <li>
            <action>HARVEST id</action>
            Make troll <var>id</var> harvest on its current cell.
          </li>

  <!-- BEGIN level2 level3 level4 -->
          <li>
  <!-- BEGIN level2 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
            <action>PLANT id type</action>
            Make troll <var>id</var> plant a <var>type</var> on its current cell: <const>PLUM</const>, <const>LEMON</const>, <const>APPLE</const> or <const>BANANA</const>.
  <!-- BEGIN level2 -->
  </span>
  <!-- END -->
          </li>
  <!-- END -->

  <!-- BEGIN level3 level4 -->
          <li>
  <!-- BEGIN level3 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
            <action>CHOP id</action>
            Make troll <var>id</var> chop on its current cell.
  <!-- BEGIN level3 -->
  </span>
  <!-- END -->
          </li>
  <!-- END -->

  <!-- BEGIN level2 level3 level4 -->
          <li>
  <!-- BEGIN level2 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
            <action>PICK id type</action>
            Make troll <var>id</var> pick one <var>type</var> from the shack: <const>PLUM</const>, <const>LEMON</const>, <const>APPLE</const> or <const>BANANA</const>.
  <!-- BEGIN level2 -->
  </span>
  <!-- END -->
          </li>
  <!-- END -->

          <li>
            <action>DROP id</action>
            Make troll <var>id</var> drop all carried items at the shack.
          </li>

  <!-- BEGIN level2 level3 level4 -->
          <li>
  <!-- BEGIN level2 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
            <action>TRAIN moveSpeed carryCapacity harvestPower chopPower</action>
            Train a new troll with the given attributes.
  <!-- BEGIN level2 -->
  </span>
  <!-- END -->
          </li>
  <!-- END -->

  <!-- BEGIN level3 level4 -->
          <li>
  <!-- BEGIN level3 -->
  <span style="color: #7cc576; background-color: rgba(124, 197, 118,.1); padding: 2px;">
  <!-- END -->
            <action>MINE id</action>
            Make troll <var>id</var> mine a nearby <const>IRON</const>.
  <!-- BEGIN level3 -->
  </span>
  <!-- END -->
          </li>
  <!-- END -->
            <li><action>WAIT</action> to do nothing.</li>
            <li><action>MSG text</action> to display a message in the replay.</li>
        </ul>

      </div>

      </div>
    </div>

    <!-- Protocol block -->
    <div class="blk">
      <div class="title">
        Constraints
      </div>
      <div class="text">
        <br> Response time for the first turn &le; <const>1000</const> ms
        <br> Response time per turn &le; <const>50</const> ms
        <br><i>Note: you will only lose the game when you exceed the time limit <const>3</const> times by at most <const>50</const> ms or once by more than that. This is to counter random delays when servers are busy, it is not advised to intentionally exceed the limits.</i>
        <!-- BEGIN level1 level2 -->
        <br> The game ends after <const>100</const> turns (for this league, later <const>300</const>)
        <br> <var>height</var> = <const>8</const> (for this league, later up to <const>11</const>)
        <!-- END -->
        <!-- BEGIN level3 level4 -->
        <br> The game ends after <const>300</const> turns, when either player can enforce a win by doing nothing or when there are no trees on the map for <const>10</const> consecutive turns.
        <br> <const>8</const> &le; <var>height</var> &le; <const>11</const>
        <!-- END -->
        <br> <var>width</var> = <const>2</const> * <var>height</var>
      </div>
    </div>
 
<br><br>
  Sprites and artwork<br>
  <a href="https://creator.nightcafe.studio/creation/fN6VKq0hdoafdkOZLZAW?ru=theMSz">MSz@nightcafe: Cover art</a>,
  <a href="https://craftpix.net/freebies/free-pixel-art-plants-for-farm/">Craftpix: Plants</a>,
  <a href="https://craftpix.net/freebies/main-characters-home-free-top-down-pixel-art-asset/">Craftpix: Decor</a>,
  <a href="https://craftpix.net/freebies/free-rocks-and-stones-top-down-pixel-art/">Craftpix: Rocks</a>,
  <a href="https://craftpix.net/freebies/free-fields-tileset-pixel-art-for-tower-defense/">Craftpix: Shack</a>,
  <a href="https://craftpix.net/freebies/top-down-crystals-pixel-art/">Craftpix: Iron</a>,
  <a href="https://craftpix.net/freebies/free-top-down-boss-character-4-direction-pack/">Craftpix: Troll</a>,
  <a href="https://www.freepik.com/free-vector/set-six-tasty-fruits_1103937.htm">freepik: fruits</a>,
  <a href="https://x.com/MaxwellVador/status/1067575489781469184">MaxwellVador@twitter: ground</a>,
  <a href="https://stockcake.com/i/pixel-lotus-pond_3171665_1535071">Stockcake: sea rose</a>,
  <a href="https://www.vecteezy.com/vector-art/56636732-pixel-art-illustration-worm-bait-pixelated-fishing-bait-earth-worm-fishing-bait-icon-pixelated-for-the-pixel-art-game-and-icon-for-website-and-game">Collaborapix Studio@vecteezy: worm</a>,
  <a href="https://eduardscarpato.itch.io/toxic-frog-animations-pixel-art-2d-free">Eduardo Scarpato@itch.io: frog</a>,
  <a href="https://www.vecteezy.com/vector-art/11433466-the-skull-and-bones-white-scraps-and-garbage-cartoon-drawn-illustration">Taras Dubov@vecteezy: fish</a>,
  <a href="https://bowpixel.itch.io/meow-cat-85-animation">bowpixel@itch.io: cat</a>,
  <a href="https://rmazanek.itch.io/bird">rmazanek@itch.io: bird</a>,
  <a href="https://jitsukoan.itch.io/turtle-pack">TheRealFusion@itch.io: turtle</a>,
  <a href="https://www.iconsdb.com/red-icons/stopwatch-icon.html">icons8@iconsdb.com: stopwatch</a>
  <br>
 </div>
</div>
