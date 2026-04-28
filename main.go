package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID        int
	Title     string
	Completed bool
}

type PageData struct {
	Tasks     []Task
	Total     int
	Completed int
}

var (
	tasks  []Task
	nextID = 1
	mutex  sync.Mutex
)

var pageTemplate = template.Must(template.New("tasks").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Student Task Tracker</title>
	<style>
		:root {
			--pink: #f05b9d;
			--pink-soft: #ffd7e8;
			--purple: #744494;
			--purple-soft: #9b73b9;
			--ink: #54306f;
			--paper: #fffdfd;
			--panel: rgba(255, 255, 255, 0.9);
		}

		* {
			box-sizing: border-box;
		}

		body {
			font-family: "Segoe UI", Arial, sans-serif;
			background:
				linear-gradient(90deg, rgba(240, 91, 157, 0.06) 1px, transparent 1px),
				linear-gradient(rgba(116, 68, 148, 0.05) 1px, transparent 1px),
				#fffefe;
			background-size: 48px 48px;
			margin: 0;
			padding: 8px;
			color: var(--ink);
		}

		button,
		input {
			font: inherit;
		}

		button {
			cursor: pointer;
		}

		.page-frame {
			min-height: calc(100vh - 16px);
			border: 5px double var(--pink);
			background:
				linear-gradient(135deg, rgba(116, 68, 148, 0.08), transparent 22%),
				linear-gradient(225deg, rgba(240, 91, 157, 0.08), transparent 22%),
				var(--paper);
			position: relative;
			overflow: hidden;
			padding: 18px 26px 24px;
		}

		.page-frame::before {
			content: "";
			position: absolute;
			inset: 24px;
			border: 2px solid var(--pink);
			pointer-events: none;
		}

		.corner {
			position: absolute;
			width: 86px;
			height: 86px;
			border-color: var(--pink);
			pointer-events: none;
			z-index: 2;
		}

		.corner::before,
		.corner::after {
			content: "";
			position: absolute;
			border-color: inherit;
		}

		.corner.top-left {
			top: 18px;
			left: 18px;
			border-top: 4px solid;
			border-left: 4px solid;
		}

		.corner.top-right {
			top: 18px;
			right: 18px;
			border-top: 4px solid;
			border-right: 4px solid;
		}

		.corner.bottom-left {
			bottom: 18px;
			left: 18px;
			border-bottom: 4px solid;
			border-left: 4px solid;
			border-color: var(--purple);
		}

		.corner.bottom-right {
			right: 18px;
			bottom: 18px;
			border-right: 4px solid;
			border-bottom: 4px solid;
			border-color: var(--purple);
		}

		.corner::before {
			inset: 16px;
			border-top: 3px solid;
			border-left: 3px solid;
		}

		.corner::after {
			inset: 32px;
			border-top: 2px solid;
			border-left: 2px solid;
		}

		.corner.top-right::before,
		.corner.top-right::after,
		.corner.bottom-right::before,
		.corner.bottom-right::after {
			border-left: 0;
			border-right: 3px solid;
		}

		.corner.bottom-left::before,
		.corner.bottom-left::after,
		.corner.bottom-right::before,
		.corner.bottom-right::after {
			border-top: 0;
			border-bottom: 3px solid;
		}

		.app-shell {
			position: relative;
			z-index: 3;
			width: min(100%, 1320px);
			margin: 0 auto;
		}

		.top-bar {
			min-height: 190px;
			display: grid;
			grid-template-columns: 1fr minmax(340px, 560px) 1fr;
			align-items: center;
			gap: 24px;
			border-bottom: 2px solid var(--pink);
			padding: 6px 34px 24px;
		}

		.nav-cluster,
		.user-cluster {
			display: flex;
			align-items: center;
			gap: 28px;
		}

		.user-cluster {
			justify-content: flex-end;
		}

		.arch-badge,
		.user-badge,
		.add-badge,
		.footer-medallion {
			position: relative;
			display: grid;
			place-items: center;
			color: var(--pink);
		}

		.arch-badge {
			width: 94px;
			height: 88px;
			border: 2px solid var(--pink);
			border-bottom: 0;
			border-radius: 48px 48px 0 0;
		}

		.user-badge {
			width: 74px;
			height: 74px;
			border: 2px solid var(--pink);
			border-radius: 50%;
		}

		.add-badge {
			width: 122px;
			height: 108px;
			border: 2px solid var(--pink);
			border-bottom: 0;
			border-radius: 62px 62px 0 0;
			flex: 0 0 auto;
		}

		.fan {
			width: 64px;
			height: 56px;
			position: relative;
			overflow: hidden;
		}

		.fan span {
			position: absolute;
			left: 50%;
			bottom: 0;
			width: 3px;
			height: 54px;
			background: currentColor;
			transform-origin: bottom center;
		}

		.fan span:nth-child(1) {
			transform: translateX(-50%) rotate(-50deg);
		}

		.fan span:nth-child(2) {
			transform: translateX(-50%) rotate(-30deg);
		}

		.fan span:nth-child(3) {
			transform: translateX(-50%) rotate(-10deg);
		}

		.fan span:nth-child(4) {
			transform: translateX(-50%) rotate(10deg);
		}

		.fan span:nth-child(5) {
			transform: translateX(-50%) rotate(30deg);
		}

		.fan span:nth-child(6) {
			transform: translateX(-50%) rotate(50deg);
		}

		.fan::after {
			content: "";
			position: absolute;
			left: 50%;
			bottom: 0;
			width: 12px;
			height: 12px;
			border: 2px solid currentColor;
			background: var(--paper);
			transform: translateX(-50%) rotate(45deg);
		}

		.dashboard-link {
			display: flex;
			align-items: center;
			gap: 14px;
			color: var(--purple);
			font-size: 16px;
			font-weight: 700;
			letter-spacing: 1px;
			text-transform: uppercase;
		}

		.grid-icon {
			width: 26px;
			height: 26px;
			display: grid;
			grid-template-columns: repeat(2, 1fr);
			gap: 6px;
		}

		.grid-icon span {
			border: 2px solid var(--purple);
		}

		.title-frame {
			position: relative;
			display: grid;
			place-items: center;
			border: 4px double var(--pink);
			padding: 26px 28px 22px;
			background: rgba(255, 255, 255, 0.72);
			box-shadow: 0 0 0 8px rgba(255, 255, 255, 0.7);
		}

		.title-frame::before,
		.title-frame::after {
			content: "";
			position: absolute;
			top: 50%;
			width: 12px;
			height: 12px;
			border: 2px solid var(--pink);
			background: var(--paper);
			transform: translateY(-50%) rotate(45deg);
		}

		.title-frame::before {
			left: 26px;
		}

		.title-frame::after {
			right: 26px;
		}

		h1 {
			margin: 0;
			font-family: Georgia, "Times New Roman", serif;
			font-size: clamp(40px, 5vw, 70px);
			font-weight: 400;
			line-height: 1;
			letter-spacing: 5px;
			text-transform: uppercase;
			text-align: center;
			color: var(--purple);
		}

		.title-fan {
			position: absolute;
			top: -72px;
			left: 50%;
			width: 116px;
			height: 88px;
			color: var(--purple);
			transform: translateX(-50%);
		}

		.welcome {
			font-size: 17px;
			color: var(--purple);
		}

		.chevron {
			width: 13px;
			height: 13px;
			border-right: 3px solid var(--purple);
			border-bottom: 3px solid var(--purple);
			transform: rotate(45deg);
			margin-top: -7px;
		}

		.add-panel,
		.tasks-panel {
			border: 2px solid var(--pink);
			background: var(--panel);
			box-shadow: inset 0 0 0 2px rgba(240, 91, 157, 0.12);
		}

		.add-panel {
			position: relative;
			margin: 44px auto 28px;
			padding: 28px 34px;
			display: grid;
			grid-template-columns: 140px 1px 1fr;
			align-items: center;
			gap: 28px;
			border-radius: 8px;
		}

		.add-panel::before,
		.add-panel::after,
		.tasks-panel::before,
		.tasks-panel::after {
			content: "";
			position: absolute;
			width: 26px;
			height: 26px;
			border-color: var(--pink);
			pointer-events: none;
		}

		.add-panel::before,
		.tasks-panel::before {
			top: 8px;
			left: 8px;
			border-top: 2px solid;
			border-left: 2px solid;
		}

		.add-panel::after,
		.tasks-panel::after {
			top: 8px;
			right: 8px;
			border-top: 2px solid;
			border-right: 2px solid;
		}

		.add-divider {
			width: 1px;
			height: 84px;
			background: var(--pink-soft);
			position: relative;
		}

		.add-divider::after {
			content: "";
			position: absolute;
			top: 50%;
			left: 50%;
			width: 12px;
			height: 12px;
			border: 2px solid var(--pink);
			background: var(--paper);
			transform: translate(-50%, -50%) rotate(45deg);
		}

		.add-form {
			display: grid;
			grid-template-columns: 1fr 250px;
			gap: 22px;
			align-items: end;
			margin: 0;
		}

		.input-group {
			display: grid;
			gap: 12px;
		}

		.input-group label,
		.section-title {
			color: var(--purple);
			font-size: 22px;
			font-weight: 700;
			letter-spacing: 2px;
			text-transform: uppercase;
		}

		input[type="text"] {
			width: 100%;
			height: 64px;
			border: 2px solid var(--purple-soft);
			border-radius: 8px;
			background: #fff;
			color: var(--ink);
			padding: 0 22px;
			font-size: 18px;
			outline: none;
		}

		input[type="text"]:focus {
			border-color: var(--pink);
			box-shadow: 0 0 0 4px rgba(240, 91, 157, 0.14);
		}

		.add-button {
			height: 64px;
			display: flex;
			justify-content: center;
			align-items: center;
			gap: 18px;
			border: 4px double #ffffff;
			background: linear-gradient(135deg, var(--purple), var(--purple-soft));
			color: white;
			font-size: 18px;
			font-weight: 700;
			letter-spacing: 2px;
			text-transform: uppercase;
			box-shadow: 0 0 0 2px var(--purple-soft);
		}

		.add-button:hover {
			background: linear-gradient(135deg, #633982, #8c60ad);
		}

		.plus-icon {
			width: 28px;
			height: 28px;
			border: 2px solid currentColor;
			border-radius: 50%;
			position: relative;
			display: flex;
			flex: 0 0 auto;
		}

		.plus-icon::before,
		.plus-icon::after {
			content: "";
			position: absolute;
			top: 50%;
			left: 50%;
			width: 14px;
			height: 2px;
			background: currentColor;
			transform: translate(-50%, -50%);
		}

		.plus-icon::after {
			transform: translate(-50%, -50%) rotate(90deg);
		}

		.tasks-panel {
			position: relative;
			border-radius: 18px;
			padding: 22px 30px 12px;
		}

		.section-heading {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 18px;
			margin-bottom: 16px;
		}

		.section-heading::before,
		.section-heading::after,
		.footer-note::before,
		.footer-note::after {
			content: "";
			width: 96px;
			height: 1px;
			background: var(--pink);
		}

		.diamond {
			width: 14px;
			height: 14px;
			border: 2px solid var(--pink);
			transform: rotate(45deg);
		}

		.section-title {
			font-size: 28px;
			margin: 0;
		}

		.task-list {
			display: grid;
			gap: 8px;
		}

		.task {
			min-height: 82px;
			display: grid;
			grid-template-columns: 1fr auto;
			align-items: center;
			gap: 18px;
			border: 2px solid rgba(240, 91, 157, 0.68);
			border-radius: 8px;
			background: rgba(255, 255, 255, 0.72);
			padding: 14px 24px;
		}

		.task-main {
			display: flex;
			align-items: center;
			gap: 24px;
			min-width: 0;
		}

		.checkbox-button,
		.checkbox-done {
			width: 40px;
			height: 40px;
			border: 3px solid var(--pink);
			border-radius: 2px;
			background: #fff;
			display: grid;
			place-items: center;
			color: #fff;
			flex: 0 0 auto;
			padding: 0;
		}

		.checkbox-button:hover {
			background: var(--pink-soft);
		}

		.checkbox-done {
			border-color: var(--purple);
			background: var(--purple);
		}

		.checkbox-done::before {
			content: "";
			width: 12px;
			height: 22px;
			border-right: 4px solid white;
			border-bottom: 4px solid white;
			transform: rotate(45deg) translate(-2px, -3px);
		}

		.task-copy {
			min-width: 0;
		}

		.task-title {
			display: block;
			color: var(--purple);
			font-size: 19px;
			font-weight: 700;
			overflow-wrap: anywhere;
		}

		.task-note {
			display: block;
			margin-top: 6px;
			color: var(--purple-soft);
			font-size: 15px;
		}

		.task.is-completed .task-title,
		.task.is-completed .task-note {
			text-decoration: line-through;
			color: #87649c;
		}

		.actions {
			display: flex;
			align-items: center;
			gap: 18px;
		}

		.status-pill {
			min-width: 116px;
			border: 2px solid var(--pink);
			border-radius: 14px;
			padding: 10px 18px;
			text-align: center;
			color: var(--pink);
			font-size: 14px;
			font-weight: 700;
		}

		.task.is-completed .status-pill {
			border-color: var(--purple-soft);
			color: var(--purple);
		}

		.delete-button {
			width: 40px;
			height: 48px;
			border: 0;
			background: transparent;
			color: var(--pink);
			padding: 0;
		}

		.delete-button:hover {
			color: #c93577;
		}

		.delete-button svg {
			width: 34px;
			height: 34px;
			display: block;
		}

		.empty {
			text-align: center;
			color: var(--purple-soft);
			border: 2px dashed var(--pink-soft);
			border-radius: 8px;
			padding: 32px;
			font-size: 18px;
		}

		.task-summary {
			display: grid;
			grid-template-columns: 1fr auto 1fr;
			align-items: center;
			gap: 18px;
			margin-top: 26px;
		}

		.task-summary::before,
		.task-summary::after {
			content: "";
			height: 1px;
			background: var(--pink);
		}

		.summary-center {
			display: flex;
			align-items: center;
			gap: 0;
		}

		.stat {
			min-width: 150px;
			border: 4px double var(--pink);
			padding: 9px 18px;
			color: var(--purple);
			font-size: 14px;
			font-weight: 800;
			text-align: center;
			text-transform: uppercase;
			letter-spacing: 1px;
		}

		.footer-medallion {
			width: 74px;
			height: 74px;
			border: 3px solid var(--pink);
			border-radius: 50%;
			background: var(--paper);
			margin: 0 -2px;
		}

		.footer-medallion .fan {
			width: 50px;
			height: 42px;
		}

		.footer-note {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 16px;
			margin: 34px 0 4px;
			color: var(--purple);
			font-weight: 700;
			letter-spacing: 2px;
			text-transform: uppercase;
		}

		.footer-note .fan {
			width: 34px;
			height: 28px;
			color: var(--pink);
		}

		@media (max-width: 980px) {
			.top-bar {
				grid-template-columns: 1fr;
				padding-top: 30px;
				text-align: center;
			}

			.nav-cluster,
			.user-cluster {
				justify-content: center;
			}

			.add-panel {
				grid-template-columns: 1fr;
				justify-items: center;
			}

			.add-divider {
				display: none;
			}

			.add-form {
				width: 100%;
				grid-template-columns: 1fr;
			}
		}

		@media (max-width: 680px) {
			.page-frame {
				padding: 14px;
			}

			.page-frame::before,
			.corner,
			.arch-badge,
			.user-badge,
			.add-badge,
			.status-pill,
			.footer-note::before,
			.footer-note::after {
				display: none;
			}

			.top-bar {
				min-height: auto;
				gap: 18px;
				padding: 18px 8px;
			}

			.nav-cluster,
			.user-cluster {
				gap: 12px;
			}

			.title-frame {
				padding: 22px 18px 18px;
			}

			.title-frame::before,
			.title-frame::after,
			.title-fan {
				display: none;
			}

			h1 {
				font-size: 36px;
				letter-spacing: 2px;
			}

			.add-panel,
			.tasks-panel {
				margin-top: 18px;
				padding: 18px;
			}

			.input-group label,
			.section-title {
				font-size: 20px;
				text-align: center;
			}

			.task {
				grid-template-columns: 1fr;
				padding: 14px;
			}

			.task-main {
				gap: 14px;
			}

			.actions {
				justify-content: flex-end;
			}

			.task-summary {
				grid-template-columns: 1fr;
			}

			.task-summary::before,
			.task-summary::after {
				display: none;
			}

			.summary-center {
				justify-content: center;
				flex-wrap: wrap;
				gap: 8px;
			}

			.footer-medallion {
				order: -1;
			}
		}
	</style>
</head>
<body>
	<div class="page-frame">
		<div class="corner top-left"></div>
		<div class="corner top-right"></div>
		<div class="corner bottom-left"></div>
		<div class="corner bottom-right"></div>

		<main class="app-shell">
			<header class="top-bar">
				<div class="nav-cluster">
					<div class="arch-badge" aria-hidden="true">
						<div class="fan">
							<span></span><span></span><span></span><span></span><span></span><span></span>
						</div>
					</div>
					<div class="dashboard-link">
						<span class="grid-icon" aria-hidden="true">
							<span></span><span></span><span></span><span></span>
						</span>
						<span>Dashboard</span>
					</div>
				</div>

				<div class="title-frame">
					<div class="fan title-fan" aria-hidden="true">
						<span></span><span></span><span></span><span></span><span></span><span></span>
					</div>
					<h1>Task Manager</h1>
				</div>

				<div class="user-cluster">
					<div class="user-badge" aria-hidden="true">
						<div class="fan">
							<span></span><span></span><span></span><span></span><span></span><span></span>
						</div>
					</div>
					<span class="welcome">Welcome, User</span>
					<span class="chevron" aria-hidden="true"></span>
				</div>
			</header>

			<section class="add-panel" aria-label="Add new task">
				<div class="add-badge" aria-hidden="true">
					<div class="fan">
						<span></span><span></span><span></span><span></span><span></span><span></span>
					</div>
				</div>
				<div class="add-divider" aria-hidden="true"></div>
				<form class="add-form" action="/add" method="POST">
					<div class="input-group">
						<label for="title">Add New Task</label>
						<input id="title" type="text" name="title" placeholder="Enter a new task..." required>
					</div>
					<button class="add-button" type="submit">
						<span>Add Task</span>
						<span class="plus-icon" aria-hidden="true"></span>
					</button>
				</form>
			</section>

			<section class="tasks-panel" aria-label="Task list">
				<div class="section-heading">
					<span class="diamond" aria-hidden="true"></span>
					<h2 class="section-title">Your Tasks</h2>
					<span class="diamond" aria-hidden="true"></span>
				</div>

				{{if .Tasks}}
					<div class="task-list">
						{{range .Tasks}}
							<article class="task {{if .Completed}}is-completed{{end}}">
								<div class="task-main">
									{{if .Completed}}
										<span class="checkbox-done" aria-label="Completed task"></span>
									{{else}}
										<form action="/complete" method="POST">
											<input type="hidden" name="id" value="{{.ID}}">
											<button class="checkbox-button" type="submit" aria-label="Mark task as completed"></button>
										</form>
									{{end}}

									<div class="task-copy">
										<span class="task-title">{{.Title}}</span>
										<span class="task-note">{{if .Completed}}Finished school task{{else}}School task to complete{{end}}</span>
									</div>
								</div>

								<div class="actions">
									<span class="status-pill">{{if .Completed}}Completed{{else}}Pending{{end}}</span>
									<form action="/delete" method="POST">
										<input type="hidden" name="id" value="{{.ID}}">
										<button class="delete-button" type="submit" aria-label="Delete task">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
												<path d="M3 6h18"></path>
												<path d="M8 6V4h8v2"></path>
												<path d="M6 6l1 16h10l1-16"></path>
												<path d="M10 10v8"></path>
												<path d="M14 10v8"></path>
											</svg>
										</button>
									</form>
								</div>
							</article>
						{{end}}
					</div>
				{{else}}
					<p class="empty">No tasks yet. Add your first task!</p>
				{{end}}

				<div class="task-summary">
					<div class="summary-center">
						<div class="stat">{{.Total}} Tasks</div>
						<div class="footer-medallion" aria-hidden="true">
							<div class="fan">
								<span></span><span></span><span></span><span></span><span></span><span></span>
							</div>
						</div>
						<div class="stat">{{.Completed}} Completed</div>
					</div>
				</div>
			</section>

			<footer class="footer-note">
				<div class="fan" aria-hidden="true">
					<span></span><span></span><span></span><span></span><span></span><span></span>
				</div>
				<span>Stay organized. Achieve more.</span>
			</footer>
		</main>
	</div>
</body>
</html>
`))

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	mutex.Lock()
	taskList := make([]Task, len(tasks))
	copy(taskList, tasks)
	mutex.Unlock()

	data := PageData{
		Tasks:     taskList,
		Total:     len(taskList),
		Completed: countCompletedTasks(taskList),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := pageTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
	}
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title != "" {
		mutex.Lock()
		tasks = append(tasks, Task{
			ID:        nextID,
			Title:     title,
			Completed: false,
		})
		nextID++
		mutex.Unlock()
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err == nil {
		mutex.Lock()
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Completed = true
				break
			}
		}
		mutex.Unlock()
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err == nil {
		mutex.Lock()
		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
		mutex.Unlock()
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func countCompletedTasks(taskList []Task) int {
	completed := 0
	for _, task := range taskList {
		if task.Completed {
			completed++
		}
	}
	return completed
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/add", addTaskHandler)
	http.HandleFunc("/complete", completeTaskHandler)
	http.HandleFunc("/delete", deleteTaskHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
