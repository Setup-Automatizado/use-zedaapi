"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

import { CreateInstanceSchema, type CreateInstanceInput } from "@/schemas";
import { createInstance } from "@/actions";
import { isError } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import {
	Field,
	FieldContent,
	FieldDescription,
	FieldError,
	FieldGroup,
	FieldLabel,
	FieldTitle,
} from "@/components/ui/field";

export function CreateInstanceForm() {
	const router = useRouter();
	const [isSubmitting, setIsSubmitting] = useState(false);

	const {
		register,
		handleSubmit,
		watch,
		setValue,
		formState: { errors },
	} = useForm({
		resolver: zodResolver(CreateInstanceSchema),
		defaultValues: {
			name: "",
			sessionName: undefined,
			isDevice: false,
			businessDevice: false,
			notifySentByMe: false,
			callRejectAuto: false,
			callRejectMessage: undefined,
			autoReadMessage: false,
			deliveryCallbackUrl: undefined,
			receivedCallbackUrl: undefined,
			receivedAndDeliveryCallbackUrl: undefined,
			messageStatusCallbackUrl: undefined,
			connectedCallbackUrl: undefined,
			disconnectedCallbackUrl: undefined,
			presenceChatCallbackUrl: undefined,
		},
	});

	const callRejectAuto = watch("callRejectAuto");

	const onSubmit = async (data: CreateInstanceInput) => {
		setIsSubmitting(true);

		try {
			const formData = new FormData();

			// Add all form fields
			formData.append("name", data.name);
			if (data.sessionName)
				formData.append("sessionName", data.sessionName);
			formData.append("isDevice", String(data.isDevice));
			formData.append("businessDevice", String(data.businessDevice));
			formData.append("notifySentByMe", String(data.notifySentByMe));
			formData.append("callRejectAuto", String(data.callRejectAuto));
			if (data.callRejectMessage)
				formData.append("callRejectMessage", data.callRejectMessage);
			formData.append("autoReadMessage", String(data.autoReadMessage));

			// Add webhook URLs
			if (data.deliveryCallbackUrl)
				formData.append(
					"deliveryCallbackUrl",
					data.deliveryCallbackUrl,
				);
			if (data.receivedCallbackUrl)
				formData.append(
					"receivedCallbackUrl",
					data.receivedCallbackUrl,
				);
			if (data.receivedAndDeliveryCallbackUrl)
				formData.append(
					"receivedAndDeliveryCallbackUrl",
					data.receivedAndDeliveryCallbackUrl,
				);
			if (data.messageStatusCallbackUrl)
				formData.append(
					"messageStatusCallbackUrl",
					data.messageStatusCallbackUrl,
				);
			if (data.connectedCallbackUrl)
				formData.append(
					"connectedCallbackUrl",
					data.connectedCallbackUrl,
				);
			if (data.disconnectedCallbackUrl)
				formData.append(
					"disconnectedCallbackUrl",
					data.disconnectedCallbackUrl,
				);
			if (data.presenceChatCallbackUrl)
				formData.append(
					"presenceChatCallbackUrl",
					data.presenceChatCallbackUrl,
				);

			const result = await createInstance(formData);

			if (isError(result)) {
				if (result.errors) {
					// Show field-level errors
					Object.values(result.errors).forEach((messages) => {
						messages.forEach((message) => toast.error(message));
					});
				} else {
					toast.error(result.error || "Error creating instance");
				}
				return;
			}

			toast.success("Instance created successfully!");

			// Redirect to instance details page
			if (result.data?.id) {
				router.push(`/instances/${result.data.id}`);
			} else {
				router.push("/instances");
			}
		} catch (error) {
			toast.error("Unexpected error creating instance");
			console.error("Create instance error:", error);
		} finally {
			setIsSubmitting(false);
		}
	};

	return (
		<form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
			{/* Basic Information */}
			<FieldGroup>
				<Field>
					<FieldLabel htmlFor="name">
						Instance Name{" "}
						<span className="text-destructive">*</span>
					</FieldLabel>
					<FieldContent>
						<Input
							id="name"
							placeholder="E.g.: WhatsApp Production"
							{...register("name")}
							disabled={isSubmitting}
						/>
						<FieldDescription>
							Identifier name for this instance
						</FieldDescription>
						{errors.name && (
							<FieldError>{errors.name.message}</FieldError>
						)}
					</FieldContent>
				</Field>

				<Field>
					<FieldLabel htmlFor="sessionName">Session Name</FieldLabel>
					<FieldContent>
						<Input
							id="sessionName"
							placeholder="E.g.: session-prod-01"
							{...register("sessionName")}
							disabled={isSubmitting}
						/>
						<FieldDescription>
							Technical session identifier (optional)
						</FieldDescription>
						{errors.sessionName && (
							<FieldError>
								{errors.sessionName.message}
							</FieldError>
						)}
					</FieldContent>
				</Field>

				<div className="grid gap-4 sm:grid-cols-2">
					<Field orientation="horizontal">
						<FieldLabel
							htmlFor="isDevice"
							className="flex items-center gap-2"
						>
							<Checkbox
								id="isDevice"
								checked={watch("isDevice")}
								onCheckedChange={(checked) =>
									setValue("isDevice", !!checked)
								}
								disabled={isSubmitting}
							/>
							<FieldTitle>Use as Device</FieldTitle>
						</FieldLabel>
						<FieldDescription>
							Connect as linked device
						</FieldDescription>
					</Field>

					<Field orientation="horizontal">
						<FieldLabel
							htmlFor="businessDevice"
							className="flex items-center gap-2"
						>
							<Checkbox
								id="businessDevice"
								checked={watch("businessDevice")}
								onCheckedChange={(checked) =>
									setValue("businessDevice", !!checked)
								}
								disabled={isSubmitting}
							/>
							<FieldTitle>WhatsApp Business</FieldTitle>
						</FieldLabel>
						<FieldDescription>
							Configure as business account
						</FieldDescription>
					</Field>
				</div>
			</FieldGroup>

			{/* Accordion Sections */}
			<Accordion type="multiple" className="w-full">
				{/* Webhooks Section */}
				<AccordionItem value="webhooks">
					<AccordionTrigger>
						Webhook Configuration (Optional)
					</AccordionTrigger>
					<AccordionContent>
						<FieldGroup className="pt-4">
							<Field>
								<FieldLabel htmlFor="deliveryCallbackUrl">
									Delivery URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="deliveryCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/delivery"
										{...register("deliveryCallbackUrl")}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Receive message delivery notifications
									</FieldDescription>
									{errors.deliveryCallbackUrl && (
										<FieldError>
											{errors.deliveryCallbackUrl.message}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="receivedCallbackUrl">
									Received URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="receivedCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/received"
										{...register("receivedCallbackUrl")}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Receive incoming messages
									</FieldDescription>
									{errors.receivedCallbackUrl && (
										<FieldError>
											{errors.receivedCallbackUrl.message}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="receivedAndDeliveryCallbackUrl">
									Combined URL (Received + Delivery)
								</FieldLabel>
								<FieldContent>
									<Input
										id="receivedAndDeliveryCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/messages"
										{...register(
											"receivedAndDeliveryCallbackUrl",
										)}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Receive both incoming messages and
										deliveries
									</FieldDescription>
									{errors.receivedAndDeliveryCallbackUrl && (
										<FieldError>
											{
												errors
													.receivedAndDeliveryCallbackUrl
													.message
											}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="messageStatusCallbackUrl">
									Message Status URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="messageStatusCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/status"
										{...register(
											"messageStatusCallbackUrl",
										)}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Receive status updates (sent, delivered,
										read)
									</FieldDescription>
									{errors.messageStatusCallbackUrl && (
										<FieldError>
											{
												errors.messageStatusCallbackUrl
													.message
											}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="connectedCallbackUrl">
									Connection URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="connectedCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/connected"
										{...register("connectedCallbackUrl")}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Notify when the instance connects
									</FieldDescription>
									{errors.connectedCallbackUrl && (
										<FieldError>
											{
												errors.connectedCallbackUrl
													.message
											}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="disconnectedCallbackUrl">
									Disconnection URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="disconnectedCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/disconnected"
										{...register("disconnectedCallbackUrl")}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Notify when the instance disconnects
									</FieldDescription>
									{errors.disconnectedCallbackUrl && (
										<FieldError>
											{
												errors.disconnectedCallbackUrl
													.message
											}
										</FieldError>
									)}
								</FieldContent>
							</Field>

							<Field>
								<FieldLabel htmlFor="presenceChatCallbackUrl">
									Chat Presence URL
								</FieldLabel>
								<FieldContent>
									<Input
										id="presenceChatCallbackUrl"
										type="url"
										placeholder="https://your-server.com/webhook/presence"
										{...register("presenceChatCallbackUrl")}
										disabled={isSubmitting}
									/>
									<FieldDescription>
										Receive presence updates (typing,
										online, etc.)
									</FieldDescription>
									{errors.presenceChatCallbackUrl && (
										<FieldError>
											{
												errors.presenceChatCallbackUrl
													.message
											}
										</FieldError>
									)}
								</FieldContent>
							</Field>
						</FieldGroup>
					</AccordionContent>
				</AccordionItem>

				{/* Settings Section */}
				<AccordionItem value="settings">
					<AccordionTrigger>
						Advanced Settings (Optional)
					</AccordionTrigger>
					<AccordionContent>
						<FieldGroup className="pt-4">
							<Field orientation="horizontal">
								<FieldLabel
									htmlFor="notifySentByMe"
									className="flex items-center gap-2"
								>
									<Checkbox
										id="notifySentByMe"
										checked={watch("notifySentByMe")}
										onCheckedChange={(checked) =>
											setValue(
												"notifySentByMe",
												!!checked,
											)
										}
										disabled={isSubmitting}
									/>
									<FieldTitle>
										Notify Messages Sent by Me
									</FieldTitle>
								</FieldLabel>
								<FieldDescription>
									Receive webhooks for messages sent by this
									instance
								</FieldDescription>
							</Field>

							<Field orientation="horizontal">
								<FieldLabel
									htmlFor="callRejectAuto"
									className="flex items-center gap-2"
								>
									<Checkbox
										id="callRejectAuto"
										checked={watch("callRejectAuto")}
										onCheckedChange={(checked) =>
											setValue(
												"callRejectAuto",
												!!checked,
											)
										}
										disabled={isSubmitting}
									/>
									<FieldTitle>
										Reject Calls Automatically
									</FieldTitle>
								</FieldLabel>
								<FieldDescription>
									Automatically reject all incoming calls
								</FieldDescription>
							</Field>

							{callRejectAuto && (
								<Field>
									<FieldLabel htmlFor="callRejectMessage">
										Call Rejection Message
									</FieldLabel>
									<FieldContent>
										<Textarea
											id="callRejectMessage"
											placeholder="Sorry, I don't accept calls at the moment."
											rows={3}
											{...register("callRejectMessage")}
											disabled={isSubmitting}
										/>
										<FieldDescription>
											Custom message sent when rejecting
											calls (max. 500 characters)
										</FieldDescription>
										{errors.callRejectMessage && (
											<FieldError>
												{
													errors.callRejectMessage
														.message
												}
											</FieldError>
										)}
									</FieldContent>
								</Field>
							)}

							<Field orientation="horizontal">
								<FieldLabel
									htmlFor="autoReadMessage"
									className="flex items-center gap-2"
								>
									<Checkbox
										id="autoReadMessage"
										checked={watch("autoReadMessage")}
										onCheckedChange={(checked) =>
											setValue(
												"autoReadMessage",
												!!checked,
											)
										}
										disabled={isSubmitting}
									/>
									<FieldTitle>
										Mark Messages as Read Automatically
									</FieldTitle>
								</FieldLabel>
								<FieldDescription>
									Automatically mark all received messages as
									read
								</FieldDescription>
							</Field>
						</FieldGroup>
					</AccordionContent>
				</AccordionItem>
			</Accordion>

			{/* Submit Button */}
			<div className="flex justify-end gap-4">
				<Button
					type="button"
					variant="outline"
					onClick={() => router.back()}
					disabled={isSubmitting}
				>
					Cancel
				</Button>
				<Button type="submit" disabled={isSubmitting}>
					{isSubmitting && (
						<Loader2 className="mr-2 h-4 w-4 animate-spin" />
					)}
					{isSubmitting ? "Creating..." : "Create Instance"}
				</Button>
			</div>
		</form>
	);
}
