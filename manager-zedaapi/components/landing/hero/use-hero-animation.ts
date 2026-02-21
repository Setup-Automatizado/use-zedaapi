"use client";

import { useCallback, useEffect, useReducer, useRef } from "react";
import { DEMO_SCENARIOS, TIMING, type MessageType } from "./animation-data";

export type AnimationPhase =
	| "idle"
	| "typing_command"
	| "awaiting_response"
	| "response_received"
	| "message_appeared"
	| "status_sent"
	| "status_delivered"
	| "status_read"
	| "pause";

export interface ChatMessage {
	id: string;
	type: MessageType;
	text?: string;
	caption?: string;
	audioDuration?: string;
	timestamp: string;
	status: "sending" | "sent" | "delivered" | "read";
}

interface AnimationState {
	scenarioIndex: number;
	phase: AnimationPhase;
	typedChars: number;
	showResponse: boolean;
	showTypingIndicator: boolean;
	chatMessages: ChatMessage[];
}

type AnimationAction =
	| { type: "TICK_CHAR" }
	| { type: "SET_PHASE"; phase: AnimationPhase }
	| { type: "SHOW_RESPONSE" }
	| { type: "SHOW_TYPING_INDICATOR" }
	| { type: "HIDE_TYPING_INDICATOR" }
	| { type: "ADD_MESSAGE"; message: ChatMessage }
	| {
			type: "UPDATE_MESSAGE_STATUS";
			id: string;
			status: ChatMessage["status"];
	  }
	| { type: "NEXT_SCENARIO" };

function getTimestamp(): string {
	const now = new Date();
	return `${now.getHours().toString().padStart(2, "0")}:${now.getMinutes().toString().padStart(2, "0")}`;
}

const MAX_MESSAGES = 5;

function reducer(
	state: AnimationState,
	action: AnimationAction,
): AnimationState {
	switch (action.type) {
		case "TICK_CHAR":
			return { ...state, typedChars: state.typedChars + 1 };
		case "SET_PHASE":
			return { ...state, phase: action.phase };
		case "SHOW_RESPONSE":
			return { ...state, showResponse: true };
		case "SHOW_TYPING_INDICATOR":
			return { ...state, showTypingIndicator: true };
		case "HIDE_TYPING_INDICATOR":
			return { ...state, showTypingIndicator: false };
		case "ADD_MESSAGE": {
			const messages = [...state.chatMessages, action.message];
			return {
				...state,
				chatMessages: messages.slice(-MAX_MESSAGES),
				showTypingIndicator: false,
			};
		}
		case "UPDATE_MESSAGE_STATUS":
			return {
				...state,
				chatMessages: state.chatMessages.map((m) =>
					m.id === action.id ? { ...m, status: action.status } : m,
				),
			};
		case "NEXT_SCENARIO":
			return {
				...state,
				scenarioIndex:
					(state.scenarioIndex + 1) % DEMO_SCENARIOS.length,
				phase: "idle",
				typedChars: 0,
				showResponse: false,
				showTypingIndicator: false,
			};
		default:
			return state;
	}
}

const initialState: AnimationState = {
	scenarioIndex: 0,
	phase: "idle",
	typedChars: 0,
	showResponse: false,
	showTypingIndicator: false,
	chatMessages: [],
};

export function useHeroAnimation(enabled: boolean) {
	const [state, dispatch] = useReducer(reducer, initialState);
	const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const charTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);
	// Ref to read chatMessages without including in useEffect deps (avoids infinite loop)
	const chatMessagesRef = useRef(state.chatMessages);
	chatMessagesRef.current = state.chatMessages;

	const clearTimers = useCallback(() => {
		if (timerRef.current) {
			clearTimeout(timerRef.current);
			timerRef.current = null;
		}
		if (charTimerRef.current) {
			clearInterval(charTimerRef.current);
			charTimerRef.current = null;
		}
	}, []);

	const scenario = DEMO_SCENARIOS[state.scenarioIndex];

	useEffect(() => {
		if (!enabled || !scenario) return;

		clearTimers();

		const commandLength = scenario.command.length;

		switch (state.phase) {
			case "idle": {
				timerRef.current = setTimeout(() => {
					dispatch({ type: "SET_PHASE", phase: "typing_command" });
				}, 400);
				break;
			}

			case "typing_command": {
				if (state.typedChars >= commandLength) {
					timerRef.current = setTimeout(() => {
						dispatch({
							type: "SET_PHASE",
							phase: "awaiting_response",
						});
					}, TIMING.responseDelay);
				} else {
					charTimerRef.current = setInterval(() => {
						dispatch({ type: "TICK_CHAR" });
					}, TIMING.charDelay);
				}
				break;
			}

			case "awaiting_response": {
				dispatch({ type: "SHOW_RESPONSE" });
				dispatch({ type: "SHOW_TYPING_INDICATOR" });
				timerRef.current = setTimeout(() => {
					dispatch({ type: "SET_PHASE", phase: "response_received" });
				}, TIMING.typingIndicatorDuration);
				break;
			}

			case "response_received": {
				const messageId = `${scenario.id}-${Date.now()}`;
				const newMessage: ChatMessage = {
					id: messageId,
					type: scenario.type,
					text: scenario.message.text,
					caption: scenario.message.caption,
					audioDuration: scenario.message.audioDuration,
					timestamp: getTimestamp(),
					status: "sending",
				};
				timerRef.current = setTimeout(() => {
					dispatch({ type: "ADD_MESSAGE", message: newMessage });
					dispatch({ type: "SET_PHASE", phase: "message_appeared" });
				}, TIMING.messageAppearDelay);
				break;
			}

			case "message_appeared": {
				timerRef.current = setTimeout(() => {
					const lastMsg =
						chatMessagesRef.current[
							chatMessagesRef.current.length - 1
						];
					if (lastMsg) {
						dispatch({
							type: "UPDATE_MESSAGE_STATUS",
							id: lastMsg.id,
							status: "sent",
						});
					}
					dispatch({ type: "SET_PHASE", phase: "status_sent" });
				}, TIMING.sentDelay);
				break;
			}

			case "status_sent": {
				timerRef.current = setTimeout(() => {
					const lastMsg =
						chatMessagesRef.current[
							chatMessagesRef.current.length - 1
						];
					if (lastMsg) {
						dispatch({
							type: "UPDATE_MESSAGE_STATUS",
							id: lastMsg.id,
							status: "delivered",
						});
					}
					dispatch({
						type: "SET_PHASE",
						phase: "status_delivered",
					});
				}, TIMING.deliveredDelay);
				break;
			}

			case "status_delivered": {
				timerRef.current = setTimeout(() => {
					const lastMsg =
						chatMessagesRef.current[
							chatMessagesRef.current.length - 1
						];
					if (lastMsg) {
						dispatch({
							type: "UPDATE_MESSAGE_STATUS",
							id: lastMsg.id,
							status: "read",
						});
					}
					dispatch({ type: "SET_PHASE", phase: "status_read" });
				}, TIMING.readDelay);
				break;
			}

			case "status_read": {
				timerRef.current = setTimeout(() => {
					dispatch({ type: "SET_PHASE", phase: "pause" });
				}, 1000);
				break;
			}

			case "pause": {
				timerRef.current = setTimeout(() => {
					dispatch({ type: "NEXT_SCENARIO" });
				}, TIMING.pauseBetweenScenarios);
				break;
			}
		}

		return clearTimers;
	}, [enabled, state.phase, state.typedChars, scenario, clearTimers]);

	// Stop char timer when we've typed the full command
	useEffect(() => {
		if (!scenario) return;
		if (
			state.typedChars >= scenario.command.length &&
			charTimerRef.current
		) {
			clearInterval(charTimerRef.current);
			charTimerRef.current = null;
		}
	}, [state.typedChars, scenario]);

	return {
		phase: state.phase,
		scenarioIndex: state.scenarioIndex,
		typedChars: state.typedChars,
		showResponse: state.showResponse,
		showTypingIndicator: state.showTypingIndicator,
		chatMessages: state.chatMessages,
		currentScenario: scenario,
	};
}
