import { z } from "zod";

/**
 * Schema for instance settings configuration.
 * Validates call rejection, message reading, and notification settings.
 */
export const InstanceSettingsSchema = z.object({
	autoReadMessage: z.boolean().default(false),

	callRejectAuto: z.boolean().default(false),

	callRejectMessage: z
		.string()
		.max(500, "Mensagem de rejeicao deve ter no maximo 500 caracteres")
		.trim()
		.optional()
		.or(z.literal(""))
		.transform((val) => (val === "" ? undefined : val)),

	notifySentByMe: z.boolean().default(false),
});

/**
 * Schema for updating instance settings.
 * All fields are optional to allow partial updates.
 */
export const UpdateSettingsSchema = z
	.object({
		autoReadMessage: z.boolean().optional(),

		callRejectAuto: z.boolean().optional(),

		callRejectMessage: z
			.string()
			.max(500, "Mensagem de rejeicao deve ter no maximo 500 caracteres")
			.trim()
			.optional()
			.or(z.literal("")),

		notifySentByMe: z.boolean().optional(),
	})
	.transform((data) => {
		// Transform empty string to undefined for callRejectMessage
		if (data.callRejectMessage === "") {
			data.callRejectMessage = undefined;
		}
		return data;
	})
	.refine(
		(data) => {
			// If callRejectAuto is true, callRejectMessage should be provided
			if (data.callRejectAuto === true && !data.callRejectMessage) {
				return false;
			}
			return true;
		},
		{
			message:
				"Mensagem de rejeicao e obrigatoria quando rejeicao automatica esta ativada",
			path: ["callRejectMessage"],
		},
	);

/**
 * Schema for profile settings.
 * Includes name, description, and profile picture configuration.
 */
export const ProfileSettingsSchema = z.object({
	profileName: z
		.string()
		.min(1, "Nome do perfil nao pode estar vazio")
		.max(25, "Nome do perfil deve ter no maximo 25 caracteres")
		.trim()
		.optional(),

	profileDescription: z
		.string()
		.max(139, "Descricao do perfil deve ter no maximo 139 caracteres")
		.trim()
		.optional()
		.or(z.literal(""))
		.transform((val) => (val === "" ? undefined : val)),

	profilePictureUrl: z
		.string()
		.url("URL da foto de perfil invalida")
		.optional()
		.or(z.literal(""))
		.transform((val) => (val === "" ? undefined : val)),
});

/**
 * Schema for privacy settings.
 * Controls visibility of profile picture, status, last seen, etc.
 */
export const PrivacySettingsSchema = z.object({
	profilePicturePrivacy: z
		.enum(["everyone", "contacts", "nobody"], {
			message: "Opcao de privacidade invalida",
		})
		.default("everyone"),

	statusPrivacy: z
		.enum(["everyone", "contacts", "nobody"], {
			message: "Opcao de privacidade invalida",
		})
		.default("everyone"),

	lastSeenPrivacy: z
		.enum(["everyone", "contacts", "nobody"], {
			message: "Opcao de privacidade invalida",
		})
		.default("everyone"),

	readReceiptsEnabled: z.boolean().default(true),
});

/**
 * Schema for message settings.
 * Configures message formatting, delivery, and archival preferences.
 */
export const MessageSettingsSchema = z.object({
	archiveChats: z.boolean().default(false),

	disappearingMessagesDuration: z
		.enum(["off", "24h", "7d", "90d"], {
			message: "Duracao invalida para mensagens temporarias",
		})
		.default("off"),

	groupNotifications: z.boolean().default(true),

	mediaAutoDownload: z
		.object({
			photos: z.boolean().default(true),
			videos: z.boolean().default(false),
			documents: z.boolean().default(false),
			audio: z.boolean().default(true),
		})
		.default({
			photos: true,
			videos: false,
			documents: false,
			audio: true,
		}),
});

/**
 * Combined settings schema.
 * Merges all setting categories into a single schema.
 */
export const AllSettingsSchema = z.object({
	instance: InstanceSettingsSchema,
	profile: ProfileSettingsSchema.optional(),
	privacy: PrivacySettingsSchema.optional(),
	messages: MessageSettingsSchema.optional(),
});

// Inferred TypeScript types
export type InstanceSettings = z.infer<typeof InstanceSettingsSchema>;
export type InstanceSettingsInput = z.input<typeof InstanceSettingsSchema>;
export type UpdateSettings = z.infer<typeof UpdateSettingsSchema>;
export type ProfileSettings = z.infer<typeof ProfileSettingsSchema>;
export type PrivacySettings = z.infer<typeof PrivacySettingsSchema>;
export type MessageSettings = z.infer<typeof MessageSettingsSchema>;
export type AllSettings = z.infer<typeof AllSettingsSchema>;
