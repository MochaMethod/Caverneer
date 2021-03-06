package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// Map describes the current state of the map, e.g., its size, tile elements, etc
type Map struct {
	areaOfMap    []int
	worldMap     [][]int8
	enemiesOnMap []*Entity
}

// Entity is the base template for players and elements
type Entity struct {
	name, entityType        string
	position                []int
	graphicChar             int8
	health, damage, defense int
	hasDied                 bool
}

// CombatEntity relates functions for use with Entities that should be able to access the combat function
type CombatEntity interface {
	attack(CombatEntity)
	takeDamage(int)
	editPosition(*Map, []int, []*Entity)
	entityOnBlankTile(*Map, []int) bool
}

// The Enemy interface is used to describe functions only able to Enemy Entity types
type Enemy interface {
	randomizeMovement(*Map, *Entity)
}

func (e *Entity) attack(target CombatEntity) {
	target.takeDamage(e.damage)
}

func (e *Entity) takeDamage(damageTaken int) {
	e.health -= (damageTaken - (e.defense / 2))
}

func createWorldMap(xLength int, yLength int) [][]int8 {
	worldMap := make([][]int8, 5)
	for i := range worldMap {
		worldMap[i] = make([]int8, 5)
	}
	return worldMap
}

func (e *Entity) combat(worldMap *Map, enemy *Entity, enemiesOnMap []*Entity) {
	playerTurn := true
	for e.health > 0 && enemy.health > 0 {
		fmt.Println("Player health: ", e.health)
		fmt.Println("Enemy health: ", enemy.health)

		if playerTurn {
			input := userInput("Please enter combat action: ", []string{"attack", "fortify"})
			if input == "attack" {
				e.attack(enemy)
				fmt.Println(strings.Title(e.name), "has attacked ", enemy.name, "for", e.damage, "points")
			} else if input == "fortify" {
				increaseHealth := float64(e.defense * 1.0 / 3)
				e.health += int(math.Round(increaseHealth))
				fmt.Println(strings.Title(e.name), "is fortifying to increase health by", increaseHealth, "points")
			}

			playerTurn = !playerTurn

		} else {
			fmt.Println(strings.Title(enemy.name), "has attacked", e.name, "for", enemy.damage, "points")
			enemy.attack(e)
			playerTurn = !playerTurn
		}
	}

	if e.health <= 0 {
		fmt.Printf("You have died...\n")
		e.hasDied = true
	}

	if enemy.health <= 0 {
		fmt.Printf("Enemy has died...\n")
		enemy.hasDied = true
	}
}

func removeEnemyIndex(a []*Entity, i int) []*Entity {
	a[i] = a[len(a)-1]
	return a[:len(a)-1]
}

func printWorldMap(worldMap *Map) {
	for i := range worldMap.worldMap {
		fmt.Println(worldMap.worldMap[i])
	}
	fmt.Printf("\n")
}

func userInput(question string, acceptableInput []string) string {
	fmt.Println(question)
	for {
		var inputList []string
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		for x := range acceptableInput {
			if strings.HasPrefix(acceptableInput[x], strings.ToLower(input)) {
				inputList = append(inputList, acceptableInput[x])
			}
		}

		// Enter just enough input to distinguish between items in the acceptableInput array.
		if len(inputList) == 0 {
			fmt.Printf("Please input another command, I do not understand.\n")
			continue

		} else if len(inputList) == 1 {
			answer := inputList[0]
			stringAnswer := string(answer)
			return stringAnswer

		} else {
			fmt.Printf("Which of the following commands did you mean?: %v \n", acceptableInput)
			continue
		}
	}
}

func (e *Entity) entityOnBlankTile(worldMap *Map, movementArray []int) bool {
	if worldMap.worldMap[e.position[0]+movementArray[0]][e.position[1]+movementArray[1]] != 0 {
		return false
	}

	return true
}

func (e *Entity) mapCollisionDetection(worldMap *Map, movementArray []int) bool {
	if e.position[0]+movementArray[0] >= worldMap.areaOfMap[0] || e.position[0]+movementArray[0] < 0 {
		return true
	}

	if e.position[1]+movementArray[1] >= worldMap.areaOfMap[1] || e.position[1]+movementArray[1] < 0 {
		return true
	}

	return false
}

func (e *Entity) editPosition(worldMap *Map, movementArray []int, enemiesOnMap []*Entity) {
	colidesWithWall := e.mapCollisionDetection(worldMap, movementArray)
	if colidesWithWall {
		if e.name == "player" {
			fmt.Printf("Cannot move this way...\n")
		}
		return
	}

	if !colidesWithWall {
		worldMap.worldMap[e.position[0]][e.position[1]] = 0
		e.position[0] += movementArray[0]
		e.position[1] += movementArray[1]
		worldMap.worldMap[e.position[0]][e.position[1]] = e.graphicChar

		for x := range enemiesOnMap {
			if e.entityType == "player" {
				if enemiesOnMap[x].position[0] == e.position[0] && enemiesOnMap[x].position[1] == e.position[1] {
					fmt.Println("Enemy: (", enemiesOnMap[x].position[0], ",", enemiesOnMap[x].position[1], ")")
					fmt.Println("E: (", e.position[0], ",", e.position[1], ")")
					e.combat(worldMap, enemiesOnMap[x], enemiesOnMap)
				}
			}
		}
	}
}

func (e *Entity) randomizeMovement(worldMap *Map) []int {
	acceptableMovement := []int{-1, 0, 1}
	x := acceptableMovement[rand.Intn(len(acceptableMovement))]
	y := acceptableMovement[rand.Intn(len(acceptableMovement))]
	movementArray := []int{x, y}
	collidesWithWall := e.mapCollisionDetection(worldMap, movementArray)
	for collidesWithWall {
		acceptableMovement := []int{-1, 0, 1}
		x := acceptableMovement[rand.Intn(len(acceptableMovement))]
		y := acceptableMovement[rand.Intn(len(acceptableMovement))]
		movementArray = []int{x, y}
		collidesWithWall = e.mapCollisionDetection(worldMap, movementArray)
	}
	entityOnBlankTile := e.entityOnBlankTile(worldMap, movementArray)
	if !entityOnBlankTile {
		e.randomizeMovement(worldMap)
	}

	return []int{x, y}
}

func createEnemies(worldMap *Map, iterationCount int) *Entity {
	enemyPosition := []int{rand.Intn(worldMap.areaOfMap[0]), rand.Intn(worldMap.areaOfMap[1])}
	fmt.Println(enemyPosition)
	enemy := &Entity{
		name:        "enemy" + strconv.Itoa(iterationCount),
		entityType:  "enemy",
		position:    enemyPosition,
		health:      10,
		damage:      5,
		defense:     2,
		graphicChar: 2,
	}

	return enemy
}

func main() {
	xLength := 5
	yLength := 5
	worldMap := &Map{
		areaOfMap:    []int{xLength, yLength},
		worldMap:     createWorldMap(xLength, yLength),
		enemiesOnMap: []*Entity{},
	}

	fmt.Printf("Adding enemy values to world map as `2`\n")
	for x := 0; x < 4; x++ {
		enemy := createEnemies(worldMap, x)
		worldMap.enemiesOnMap = append(worldMap.enemiesOnMap, enemy)
	}

	player := &Entity{
		name:        "player",
		entityType:  "player",
		position:    []int{0, 0},
		health:      10,
		damage:      6,
		defense:     10,
		graphicChar: 1,
	}

	for x := 0; x < len(worldMap.enemiesOnMap); x++ {
		fmt.Println("ENEMY NAME:", worldMap.enemiesOnMap[x].name)
		enemyPos := worldMap.enemiesOnMap[x].randomizeMovement(worldMap)
		worldMap.enemiesOnMap[x].editPosition(worldMap, enemyPos, worldMap.enemiesOnMap)
	}

	fmt.Printf("Adding player value to world map as `1`, at pos (0, 0)\n")
	player.editPosition(worldMap, []int{0, 0}, worldMap.enemiesOnMap)

	fmt.Printf("Starting loop...\n")
	fmt.Printf("\n")
	printWorldMap(worldMap)
	for {
		if player.hasDied {
			os.Exit(0)
		}

		for x := 0; x < len(worldMap.enemiesOnMap); x++ {
			if worldMap.enemiesOnMap[x].hasDied {
				worldMap.worldMap[worldMap.enemiesOnMap[x].position[0]][worldMap.enemiesOnMap[x].position[1]] = player.graphicChar
				worldMap.enemiesOnMap = removeEnemyIndex(worldMap.enemiesOnMap, x)
				fmt.Println("Combat", len(worldMap.enemiesOnMap))

			} else {
				enemyPos := worldMap.enemiesOnMap[x].randomizeMovement(worldMap)
				worldMap.enemiesOnMap[x].editPosition(worldMap, enemyPos, worldMap.enemiesOnMap)
			}
		}

		input := userInput("Please enter direction to move: ", []string{"north", "south", "east", "west"})
		if input == "north" {
			player.editPosition(worldMap, []int{-1, 0}, worldMap.enemiesOnMap)
		} else if input == "south" {
			player.editPosition(worldMap, []int{1, 0}, worldMap.enemiesOnMap)
		} else if input == "east" {
			player.editPosition(worldMap, []int{0, 1}, worldMap.enemiesOnMap)
		} else if input == "west" {
			player.editPosition(worldMap, []int{0, -1}, worldMap.enemiesOnMap)
		}

		printWorldMap(worldMap)
	}
}
