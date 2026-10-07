/*
Copyright © 2026 Achno <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/Achno/gowall/config"
	"github.com/Achno/gowall/internal/backends/upload"
	"github.com/Achno/gowall/internal/image"
	imageio "github.com/Achno/gowall/internal/image_io"
	"github.com/Achno/gowall/utils"
	"github.com/spf13/cobra"
)

func BuildUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload [INPUT]",
		Short: "Uploads images to freeimage.host or imgbb and prints the links",
		Long:  `Uploads images to freeimage.host or imgbb and prints the links`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ValidateParseUploadCmd(cmd, shared, args)
		},
		Run: RunUploadCmd,
	}

	var method string
	cmd.Flags().StringVarP(&method, "method", "m", upload.DefaultUploader, "Where to upload. Available methods: "+strings.Join(upload.Names(), ", "))

	addFlags(cmd).WithBatch().WithDir()

	return cmd
}

func RunUploadCmd(cmd *cobra.Command, args []string) {
	imageOps, err := imageio.DetermineImageOperations(shared, args, cmd)
	utils.HandleError(err, "Error")

	method, err := cmd.Flags().GetString("method")
	utils.HandleError(err, "Error")

	uploader, err := upload.New(method, upload.ImageUploadOptions{
		FreeImageAPIKey: config.GowallConfig.EnvConfig.FREEIMAGE_API_KEY,
		ImgBBAPIKey:     config.GowallConfig.EnvConfig.IMGBB_API_KEY,
	})
	utils.HandleError(err, "Error")

	err = image.UploadImgs(uploader, imageOps)
	utils.HandleError(err, "Error")
}

func ValidateParseUploadCmd(cmd *cobra.Command, flags config.GlobalSubCommandFlags, args []string) error {
	if err := validateInput(flags, args); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(BuildUploadCmd())
}
