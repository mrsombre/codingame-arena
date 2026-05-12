import { WIDTH, HEIGHT } from './core/constants.js'
import { api as entityModule } from './entity-module/GraphicEntityModule.js'
import { EntityFactory } from './entity-module/EntityFactory.js'
import { ContainerBasedEntity } from './entity-module/ContainerBasedEntity.js'

export class ExplosionModule {
  hiding = true
  initialized = false
  animatedExplosion = false
  animatedFrog = false
  animatedFish = false
  animatedBird = false
  animatedTurtle = false
  constructor(assets) {
    this.hiding = Math.random() > 0.01;
  }

  static get moduleName() {
    return 'explosionmodule'
  }
  static get dependencies() {
    return ['entitymodule']
  }

  updateScene(previousData, currentData, progress) {
    if (this.initialized || progress == 1) return
    this.initialized = true

    let explosion = entityModule.entities.get(2)
    let container = entityModule.entities.get(1)
    if (explosion != null && !this.animatedExplosion) {
        explosion.container.interactive = true
        explosion.interactive = true

        explosion.container.mouseover = () => {
          if (this.animatedExplosion) return
          this.animatedExplosion = true
          entityModule.entities.delete(container.id)
          entityModule.entities.delete(explosion.id)
          const images = explosion.currentState.images.split(',')
          for (let i = 1; i < images.length; i++) {
            (function (i) {
              setTimeout(function () {
                explosion.graphics.texture = PIXI.Texture.from(images[i])
              }, 200 * (i - 1));
            })(i);
          }
        }
    }

    let frog = entityModule.entities.get(8)
    let frogContainer = entityModule.entities.get(3)
    if (frog != null) {
        frog.container.interactive = true
        frog.interactive = true

        frog.container.click = (ev) => {
          if (this.animatedFrog) return
          this.animatedFrog = true
          entityModule.entities.delete(frogContainer.id)
          entityModule.entities.delete(frog.id)
          const images = frog.currentState.images.split(',')
          for (let i = 1; i < images.length; i++) {
            (function (i) {
              setTimeout(function () {
                frog.graphics.texture = PIXI.Texture.from(images[i])
              }, 200 * (i - 1));
            })(i);
          }
        }
    }

    let fish = entityModule.entities.get(9)
    let fishContainer = entityModule.entities.get(4)
    if (fish != null) {
        fish.container.interactive = true
        fish.interactive = true

        fish.container.click = (ev) => {
          if (this.animatedFish) return
          this.animatedFish = true
          entityModule.entities.delete(fishContainer.id)
          entityModule.entities.delete(fish.id)
          const images = fish.currentState.images.split(',')
          for (let i = 1; i < images.length; i++) {
            (function (i) {
              setTimeout(function () {
                fish.graphics.texture = PIXI.Texture.from(images[i])
              }, 200 * (i - 1));
            })(i);
          }
        }
    }

    let bird = entityModule.entities.get(10)
    let birdContainer = entityModule.entities.get(5)
    let cat = entityModule.entities.get(11)
    let catContainer = entityModule.entities.get(6)
    if (bird != null) {
        bird.container.interactive = true
        bird.interactive = true
        cat.container.interactive = true
        cat.interactive = true

        cat.container.click = bird.container.click = (ev) => {
          if (this.animatedBird) return
          this.animatedBird = true
          entityModule.entities.delete(birdContainer.id)
          entityModule.entities.delete(bird.id)
          entityModule.entities.delete(catContainer.id)
          entityModule.entities.delete(cat.id)
          const birdAnim = ["b33", "b34", "b35", "b0", "b1", "b2", "b3", "b4", "b5", "b6", "b7", "b8"]
          for (let i = 1; i < birdAnim.length; i++) {
            (function (i) {
              setTimeout(function () {
                bird.graphics.texture = PIXI.Texture.from(birdAnim[i])
              }, 200 * (i - 1));
            })(i);
          }
          const catAnim = ["c50", "c51", "c52", "c53", "c54", "c55", "c56", "c10", "c11", "c12", "c13", "c14", "c15", "c16", "c17", "c18"]
          for (let i = 1; i < catAnim.length; i++) {
            (function (i) {
              setTimeout(function () {
                cat.graphics.texture = PIXI.Texture.from(catAnim[i])
              }, 200 * (i - 1));
            })(i);
          }
        }
    }

    let turtle = entityModule.entities.get(12)
    let turtleContainer = entityModule.entities.get(7)
    if (turtle != null) {
        turtle.container.interactive = true
        turtle.interactive = true

        turtle.container.click = (ev) => {
          if (this.animatedTurtle) return
          this.animatedTurtle = true
          entityModule.entities.delete(turtleContainer.id)
          entityModule.entities.delete(turtle.id)
          const images = turtle.currentState.images.split(',')
          for (let i = 1; i < images.length; i++) {
            (function (i) {
              setTimeout(function () {
                turtle.graphics.texture = PIXI.Texture.from(images[i])
              }, 200 * (i - 1));
            })(i);
          }
        }
    }
  }

  handleFrameData(frameInfo, data = []) { }

  reinitScene(container) {
    this.initialized = false

    let explosion = entityModule.entities.get(2)
    if (explosion != null && this.hiding) {
      explosion.graphics.visible = false
      let container = entityModule.entities.get(1)
      entityModule.entities.delete(container.id)
      entityModule.entities.delete(explosion.id)
    }
  }
}