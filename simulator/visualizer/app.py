from flask import Flask, render_template
import argparse
import json
import sys
from typing import Dict, List

app = Flask(__name__, static_folder="./assets", template_folder="./templates")


class Cell:
    def __init__(self, snake=False, head=False, dead = False, food=False, hazard=False, index=0):
        self.index = index
        self.snake = snake
        self.head = head
        self.food = food
        self.dead = dead


states: List[Dict] = []


def error(msg):
    return "<p>%s<\p>" % msg


@app.route("/render/<turn>")
def get_state(turn):
    try:
        state_i = int(turn)
        if state_i < 0 or state_i >= len(states):
            raise ValueError()
        state = states[state_i]
        prev_state = None
        if state_i > 0:
            prev_state = states[state_i - 1]


    except ValueError:
        return error(
            "invlid state, must be an integer between (inclusive) 0 and %d"
            % (len(states) - 1)
        )

    board = state["board"]
    table = [[Cell() for _ in range(board["width"])] for _ in range(board["width"])]
    _, height = board["width"], board["height"]
    snakes = board.get("snakes") or []
    snake_indices = {snake["id"]: i for i, snake in enumerate(sorted(states[0]["board"]["snakes"], key=lambda x: x["id"]))}
    print(snake_indices)
    for i, snake in enumerate(snakes):
        for j, body in enumerate(snake["body"]):
            cell = table[height - 1 - body["y"]][body["x"]]
            cell.snake = True
            cell.index = snake_indices[snake["id"]]
            if j == 0:
                cell.head = True
    if prev_state:
        for i, snake in enumerate(prev_state["board"]["snakes"]):
            if snake["id"] in [s["id"] for s in snakes]:
                continue
            for j, body in enumerate(snake["body"]):
                cell = table[height - 1 - body["y"]][body["x"]]
                cell.snake = True
                cell.index = snake_indices[snake["id"]]
                cell.dead = True
                if j == 0:
                    cell.head = True
    for food in board.get("food") or []:
        cell = table[height - 1 - food["y"]][food["x"]]
        cell.food = True
    for hazard in board.get("hazards") or []:
        cell = table[height - 1 - hazard["y"]][hazard["x"]]
        cell.food = True
    return render_template(
        "index.html",
        rows=table,
        next=state_i + 1 if state_i < (len(states) - 1) else None,
        prev=state_i - 1 if state_i > 0 else None,
        rawstate=json.dumps(state, indent=2)
    )


@app.route("/")
def index():
    return get_state(0)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "states_file", type=argparse.FileType("r"), help="filename to states"
    )
    args = parser.parse_args()
    states = json.loads(args.states_file.read())
    print("loaded states file:\n%s" % states, file=sys.stderr)
    app.run("0.0.0.0", 5000)