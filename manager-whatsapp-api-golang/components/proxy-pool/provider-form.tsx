"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
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
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
	CreateProviderSchema,
	type CreateProviderFormValues,
	type CreateProviderInput,
} from "@/schemas/pool";

interface ProviderFormProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSubmit: (data: CreateProviderFormValues) => Promise<void>;
}

export function ProviderForm({
	open,
	onOpenChange,
	onSubmit,
}: ProviderFormProps) {
	const [isPending, startTransition] = useTransition();

	const form = useForm<
		CreateProviderInput,
		unknown,
		CreateProviderFormValues
	>({
		resolver: zodResolver(CreateProviderSchema),
		defaultValues: {
			name: "",
			providerType: "webshare",
			enabled: true,
			priority: 100,
			apiKey: "",
			apiEndpoint: "",
			maxProxies: 0,
			maxInstancesPerProxy: 1,
			countryCodes: [],
			rateLimitRpm: 60,
		},
	});

	const handleSubmit = (values: CreateProviderFormValues) => {
		startTransition(async () => {
			try {
				await onSubmit(values);
				form.reset();
				onOpenChange(false);
				toast.success("Provider created");
			} catch {
				toast.error("Failed to create provider");
			}
		});
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-lg">
				<DialogHeader>
					<DialogTitle>Add Proxy Provider</DialogTitle>
					<DialogDescription>
						Configure a new proxy provider to source proxies for
						your instances.
					</DialogDescription>
				</DialogHeader>

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(handleSubmit)}
						className="space-y-4"
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Name</FormLabel>
									<FormControl>
										<Input
											placeholder="My Webshare"
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="providerType"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Provider Type</FormLabel>
									<Select
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select type" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											<SelectItem value="webshare">
												Webshare
											</SelectItem>
											<SelectItem value="custom">
												Custom
											</SelectItem>
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="apiKey"
							render={({ field }) => (
								<FormItem>
									<FormLabel>API Key</FormLabel>
									<FormControl>
										<Input
											type="password"
											placeholder="Enter API key"
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="apiEndpoint"
							render={({ field }) => (
								<FormItem>
									<FormLabel>
										API Endpoint (optional)
									</FormLabel>
									<FormControl>
										<Input
											placeholder="https://proxy.webshare.io/api/v2"
											{...field}
										/>
									</FormControl>
									<FormDescription>
										Leave blank for default endpoint
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="priority"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Priority</FormLabel>
										<FormControl>
											<Input
												type="number"
												{...field}
												onChange={(e) =>
													field.onChange(
														Number(e.target.value),
													)
												}
											/>
										</FormControl>
										<FormDescription>
											Lower = preferred
										</FormDescription>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="maxInstancesPerProxy"
								render={({ field }) => (
									<FormItem>
										<FormLabel>
											Max Instances/Proxy
										</FormLabel>
										<FormControl>
											<Input
												type="number"
												{...field}
												onChange={(e) =>
													field.onChange(
														Number(e.target.value),
													)
												}
											/>
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>

						<FormField
							control={form.control}
							name="enabled"
							render={({ field }) => (
								<FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
									<div>
										<FormLabel>Enabled</FormLabel>
										<FormDescription>
											Start syncing proxies immediately
										</FormDescription>
									</div>
									<FormControl>
										<Switch
											checked={field.value}
											onCheckedChange={field.onChange}
										/>
									</FormControl>
								</FormItem>
							)}
						/>

						<DialogFooter>
							<Button
								type="button"
								variant="outline"
								onClick={() => onOpenChange(false)}
							>
								Cancel
							</Button>
							<Button type="submit" disabled={isPending}>
								{isPending && (
									<Loader2 className="mr-2 h-4 w-4 animate-spin" />
								)}
								Create Provider
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
