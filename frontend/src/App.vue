<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gray-100">
    <h1 class="text-5xl text-red-500 font-bold mb-4">TicTacToe</h1>
    <div class="grid grid-cols-3 gap-2">
      <button
        v-for="(cell, index) in board"
        :key="index"
        class="w-20 h-20 text-3xl bg-red shadow rounded flex items-center justify-center"
        @click="makeMove(index)"
      >
        {{ cell }}
      </button>
    </div>
    <p class="mt-4 text-lg">{{ statusMessage }}</p>
    <button @click="newGame" class="mt-4 px-4 py-2 bg-blue-500 text-white rounded">New Game</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'

const board = ref(Array(9).fill(''))
const statusMessage = ref('Click a cell to start playing!')
const gameId = ref(null)
const currentPlayer = ref('X')

const newGame = async () => {
  try {
    const res = await axios.post('http://localhost:8080/create-game', {
      player_x: 'Player X',
      player_o: 'Player O',
    })
    gameId.value = res.data.game_id
    board.value = Array(9).fill('')
    currentPlayer.value = 'X'
    statusMessage.value = 'New game started! Player X goes first.'
  } catch (err) {
    console.error(err)
    statusMessage.value = 'Failed to create game.'
  }
}

const makeMove = async (index) => {
  if (!gameId.value) {
    statusMessage.value = 'Please start a new game first.'
    return
  }
  const row = Math.floor(index / 3)
  const col = index % 3
  try {
    const res = await axios.post('http://localhost:8080/make-move', {
      game_id: gameId.value,
      player: currentPlayer.value,
      row,
      col,
    })
    board.value = res.data.board
    if (res.data.status === 'ongoing') {
      currentPlayer.value = currentPlayer.value === 'X' ? 'O' : 'X'
      statusMessage.value = `Player ${currentPlayer.value}'s turn`
    } else if (res.data.status === 'draw') {
      statusMessage.value = 'Game ended in a draw!'
    } else {
      statusMessage.value = `Player ${res.data.status[0]} wins!`
    }
  } catch (err) {
    console.error(err)
    statusMessage.value = 'Invalid move or server error.'
  }
}
</script>


