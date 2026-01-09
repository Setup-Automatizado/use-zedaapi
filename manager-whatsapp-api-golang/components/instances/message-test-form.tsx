/**
 * Message Test Form Component
 *
 * Complete form to test sending text messages with all FUNNELCHAT parameters:
 * - phone (required)
 * - message (required)
 * - delayMessage (optional, 1-15 seconds)
 * - delayTyping (optional, 1-15 seconds)
 * - editMessageId (optional, for editing existing messages)
 *
 * @example
 * ```tsx
 * <MessageTestForm
 *   instanceId={instance.id}
 *   instanceToken={instance.instanceToken}
 *   clientToken={clientToken}
 * />
 * ```
 */

"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Send } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

const formSchema = z.object({
	phone: z
		.string()
		.min(10, "Phone must have at least 10 digits")
		.max(15, "Phone must have at most 15 digits")
		.regex(/^[0-9]+$/, "Phone must contain only numbers (no formatting)"),
	message: z
		.string()
		.min(1, "Message is required")
		.max(4096, "Message must be at most 4096 characters"),
	delayMessage: z
		.number()
		.min(0, "Delay must be at least 0 seconds")
		.max(15, "Delay must be at most 15 seconds")
		.optional(),
	delayTyping: z
		.number()
		.min(0, "Typing delay must be at least 0 seconds")
		.max(15, "Typing delay must be at most 15 seconds")
		.optional(),
	editMessageId: z.string().optional(),
});

type FormValues = z.infer<typeof formSchema>;

export interface MessageTestFormProps {
	instanceId: string;
	instanceToken: string;
}

interface SendTextResponse {
	zaapId?: string;
	messageId?: string;
	id?: string;
	error?: string;
	message?: string;
}

export function MessageTestForm({
	instanceId,
	instanceToken,
}: MessageTestFormProps) {
	const [isSending, setIsSending] = useState(false);
	const [lastResponse, setLastResponse] = useState<SendTextResponse | null>(
		null,
	);

	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			phone: "",
			message: "",
			delayMessage: 3,
			delayTyping: 0,
			editMessageId: "",
		},
	});

	const onSubmit = async (values: FormValues) => {
		setIsSending(true);
		setLastResponse(null);

		try {
			// Build request body with instance credentials and message data
			const body: Record<string, string | number> = {
				instanceId,
				instanceToken,
				phone: values.phone,
				message: values.message,
			};

			if (values.delayMessage && values.delayMessage > 0) {
				body.delayMessage = values.delayMessage;
			}

			if (values.delayTyping && values.delayTyping > 0) {
				body.delayTyping = values.delayTyping;
			}

			if (values.editMessageId && values.editMessageId.trim() !== "") {
				body.editMessageId = values.editMessageId.trim();
			}

			// Use Next.js API route as proxy to avoid CORS issues
			const response = await fetch("/api/send-text", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(body),
			});

			const data: SendTextResponse = await response.json();

			setLastResponse(data);

			if (response.ok) {
				toast.success("Message sent successfully!", {
					description: `Message ID: ${data.messageId || data.id || "N/A"}`,
				});

				// Reset form on success
				form.reset();
			} else {
				toast.error("Failed to send message", {
					description: data.error || data.message || "Unknown error",
				});
			}
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : "Unknown error occurred";

			toast.error("Error sending message", {
				description: errorMessage,
			});

			setLastResponse({
				error: errorMessage,
			});
		} finally {
			setIsSending(false);
		}
	};

	return (
		<Card>
			<CardHeader>
				<CardTitle>Test Message Sending</CardTitle>
				<CardDescription>
					Send a test message with all available parameters (delayMessage,
					delayTyping, etc.)
				</CardDescription>
			</CardHeader>
			<CardContent>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						{/* Phone Field */}
						<FormField
							control={form.control}
							name="phone"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Phone Number *</FormLabel>
									<FormControl>
										<Input
											placeholder="5511999999999"
											{...field}
											disabled={isSending}
										/>
									</FormControl>
									<FormDescription>
										Country code + area code + number (numbers only, no
										formatting)
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						{/* Message Field */}
						<FormField
							control={form.control}
							name="message"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Message *</FormLabel>
									<FormControl>
										<Textarea
											placeholder="Welcome to *WhatsApp API*! 🚀"
											className="min-h-[100px]"
											{...field}
											disabled={isSending}
										/>
									</FormControl>
									<FormDescription>
										Supports WhatsApp formatting (*bold*, _italic_, ~strike~) and
										emojis
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						{/* Delay Message */}
						<FormField
							control={form.control}
							name="delayMessage"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Delay Before Sending (seconds)</FormLabel>
									<FormControl>
										<Input
											type="number"
											min={0}
											max={15}
											{...field}
											onChange={(e) => {
												const value = e.target.value;
												field.onChange(value === "" ? undefined : Number(value));
											}}
											value={field.value ?? ""}
											disabled={isSending}
										/>
									</FormControl>
									<FormDescription>
										Wait time before sending (0-15 seconds). Default: 1-3 seconds
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						{/* Delay Typing */}
						<FormField
							control={form.control}
							name="delayTyping"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Typing Indicator Duration (seconds)</FormLabel>
									<FormControl>
										<Input
											type="number"
											min={0}
											max={15}
											{...field}
											onChange={(e) => {
												const value = e.target.value;
												field.onChange(value === "" ? undefined : Number(value));
											}}
											value={field.value ?? ""}
											disabled={isSending}
										/>
									</FormControl>
									<FormDescription>
										Show &quot;Typing...&quot; status for this duration (0-15
										seconds). Default: 0
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						{/* Edit Message ID */}
						<FormField
							control={form.control}
							name="editMessageId"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Edit Message ID (Optional)</FormLabel>
									<FormControl>
										<Input
											placeholder="D241XXXX732339502B68"
											{...field}
											disabled={isSending}
										/>
									</FormControl>
									<FormDescription>
										If provided, edits an existing message instead of sending a
										new one
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						<Button type="submit" disabled={isSending} className="w-full">
							{isSending ? (
								<>
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
									Sending...
								</>
							) : (
								<>
									<Send className="mr-2 h-4 w-4" />
									Send Test Message
								</>
							)}
						</Button>
					</form>
				</Form>

				{/* Response Display */}
				{lastResponse && (
					<div className="mt-6 rounded-lg border p-4">
						<h4 className="mb-2 font-semibold">Last Response:</h4>
						<pre className="overflow-x-auto rounded bg-muted p-2 text-xs">
							{JSON.stringify(lastResponse, null, 2)}
						</pre>
					</div>
				)}
			</CardContent>
		</Card>
	);
}
