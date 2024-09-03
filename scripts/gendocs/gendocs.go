package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kyma-project/modulectl/cmd/modulectl"
)

const (
	docsTargetDir = "./docs/gen-docs"
	fmTemplate    = `---
title: %s
---

`
)

func main() {
	command, err := modulectl.NewCmd()
	if err != nil {
		fmt.Println("unable to generate docs", err.Error())
		os.Exit(1)
	}

	err = genMarkdownTree(command, docsTargetDir)
	if err != nil {
		fmt.Println("unable to generate docs", err.Error())
		os.Exit(1)
	}

	fmt.Println("Docs successfully generated to the following dir", docsTargetDir)
	os.Exit(0)
}

func genMarkdownTree(cmd *cobra.Command, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTree(c, dir); err != nil {
			return err
		}
	}

	basename := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
	filename := filepath.Join(dir, basename)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer file.Close()

	if _, err = io.WriteString(file, filePrepender(cmd)); err != nil {
		return fmt.Errorf("writing to file failed: %w", err)
	}
	return genMarkdown(cmd, file)
}

func genMarkdown(cmd *cobra.Command, writer io.Writer) error {
	cmd.InitDefaultHelpCmd()
	initCustomHelpFlag(cmd)

	buf := new(bytes.Buffer)

	printShort(buf, cmd)
	printSynopsis(buf, cmd)

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}

	printSeeAlso(buf, cmd)

	_, err := buf.WriteTo(writer)
	if err != nil {
		return fmt.Errorf("buffer write failed: %w", err)
	}
	return nil
}

func printShort(buf *bytes.Buffer, cmd *cobra.Command) {
	short := cmd.Short
	buf.WriteString(short + "\n\n")
}

func printSynopsis(buf *bytes.Buffer, cmd *cobra.Command) {
	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("## Synopsis\n\n")
	buf.WriteString(long + "\n\n")
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Flags\n\n```bash\n")
		printFlagsWithOnlyUsage(buf, cmd)
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Flags inherited from parent commands\n\n```bash\n")
		printFlagsWithOnlyUsage(buf, cmd)
		buf.WriteString("```\n\n")
	}

	return nil
}

func printFlagsWithOnlyUsage(buf *bytes.Buffer, cmd *cobra.Command) {
	// Calculate the maximum length of the flag names (shorthand + long name)
	maxLength := 0
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		uniformSpacingFactor := 6
		flagLength := len(flag.Shorthand) + len(flag.Name) + uniformSpacingFactor // 6 accounts for the formatting "-s, --"
		if flagLength > maxLength {
			maxLength = flagLength
		}
	})

	// Print the flags with uniform spacing
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		numberOfSpacesAfterShort := 4
		// Format the flag name
		flagShort := strings.Repeat(" ", numberOfSpacesAfterShort)
		if flag.Shorthand != "" {
			flagShort = fmt.Sprintf("-%s, ", flag.Shorthand)
		}
		flagType := flag.Value.Type()
		if flagType == "bool" {
			flagType = ""
		}
		flagName := fmt.Sprintf("%s--%s %s ", flagShort, flag.Name, flagType)

		// Calculate padding to align descriptions
		paddingFactor := 10
		padding := strings.Repeat(" ", maxLength-len(flagName)+paddingFactor)

		// Print the flag name with its usage
		customString := fmt.Sprintf("%s%s%s\n", flagName, padding, flag.Usage)
		buf.WriteString(customString)
	})
}

func printSeeAlso(buf *bytes.Buffer, cmd *cobra.Command) {
	if !hasSeeAlso(cmd) {
		return
	}

	name := cmd.CommandPath()

	buf.WriteString("## See also\n\n")
	if cmd.HasParent() {
		parent := cmd.Parent()
		pname := parent.CommandPath()
		buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, linkHandler(parent), parent.Short))
		cmd.VisitParents(func(c *cobra.Command) {
			if c.DisableAutoGenTag {
				cmd.DisableAutoGenTag = c.DisableAutoGenTag
			}
		})
	}

	children := cmd.Commands()
	sort.Sort(byName(children))

	for _, child := range children {
		if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
			continue
		}
		cname := name + " " + child.Name()
		buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, linkHandler(child), child.Short))
	}
	buf.WriteString("\n")
}

func filePrepender(cmd *cobra.Command) string {
	name := cmd.CommandPath()
	return fmt.Sprintf(fmTemplate, name)
}

func linkHandler(cmd *cobra.Command) string {
	name := cmd.CommandPath()
	formatted := strings.ReplaceAll(name, " ", "_")
	return formatted + ".md"
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

func initCustomHelpFlag(cmd *cobra.Command) {
	if cmd.Flags().Lookup("help") == nil {
		usage := "Provides help for the "
		name := cmd.Name()
		if name == "" {
			usage += "this command"
		} else {
			usage += name
		}
		cmd.Flags().BoolP("help", "h", false, usage+" command.")
	}
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
