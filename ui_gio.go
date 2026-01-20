package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/sqweek/dialog"
)

type LauncherUI struct {
	theme *material.Theme

	usernameField    widget.Editor
	installPathField widget.Editor

	playBtn      widget.Clickable
	playDropdown widget.Clickable
	dropdownOpen bool
	deleteBtn    widget.Clickable
	browseBtn    widget.Clickable
	closeBtn     widget.Clickable
	minimizeBtn  widget.Clickable
	maximizeBtn  widget.Clickable

	patchlineRelease    widget.Clickable
	patchlinePreRelease widget.Clickable

	versionButtons []widget.Clickable

	modeOffline    widget.Clickable
	modeFakeOnline widget.Clickable

	decorations      widget.Decorations
	titleBarTag      bool
	lastTitlePress   int64
	normalWindowSize image.Point
}

func NewLauncherUI() *LauncherUI {
	ui := &LauncherUI{
		theme: material.NewTheme(),
	}
	ui.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	ui.usernameField.SingleLine = true
	ui.usernameField.SetText(wCommune.Username)
	ui.installPathField.SingleLine = true
	ui.installPathField.SetText(wCommune.GameFolder)

	ui.updateVersionButtons()

	return ui
}

func (ui *LauncherUI) updateVersionButtons() {
	latest := wCommune.LatestVersions[wCommune.Patchline]
	if len(ui.versionButtons) != latest {
		ui.versionButtons = make([]widget.Clickable, latest)
	}
}

type C = layout.Context
type D = layout.Dimensions

func (ui *LauncherUI) Layout(gtx C, window *app.Window) D {
	if txt := ui.usernameField.Text(); txt != wCommune.Username {
		wCommune.Username = txt
	}

	if txt := ui.installPathField.Text(); txt != wCommune.GameFolder {
		wCommune.GameFolder = txt
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return ui.layoutTitleBar(gtx, window)
		}),
		layout.Flexed(1, func(gtx C) D {
			return ui.layoutCenterArea(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return ui.layoutBottomBar(gtx)
		}),
	)
}

func (ui *LauncherUI) layoutTitleBar(gtx C, window *app.Window) D {
	height := gtx.Dp(40)
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	rect := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: height}}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 00, G: 00, B: 00, A: 255}, rect.Op())

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx C) D {
			areaSize := image.Point{X: gtx.Constraints.Max.X, Y: height}

			layout.Inset{
				Left: unit.Dp(16),
				Top:  unit.Dp(10),
			}.Layout(gtx, func(gtx C) D {
				lbl := material.Label(ui.theme, unit.Sp(15), "hytLauncher")
				lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				lbl.Font.Weight = font.Bold
				return lbl.Layout(gtx)
			})

			area := clip.Rect{Max: areaSize}.Push(gtx.Ops)

			if ui.decorations.Maximized {
				for {
					ev, ok := gtx.Event(pointer.Filter{
						Target: &ui.titleBarTag,
						Kinds:  pointer.Press,
					})
					if !ok {
						break
					}

					if pe, ok := ev.(pointer.Event); ok && pe.Kind == pointer.Press {
						now := gtx.Now.UnixMilli()
						if now-ui.lastTitlePress < 400 {
							window.Perform(system.ActionUnmaximize)
							ui.lastTitlePress = 0
						} else {
							ui.lastTitlePress = now

							normalWidth := ui.normalWindowSize.X
							if normalWidth == 0 {
								normalWidth = gtx.Dp(900)
							}

							window.Perform(system.ActionUnmaximize)
							window.Option(app.Size(unit.Dp(float32(normalWidth)/gtx.Metric.PxPerDp), unit.Dp(float32(ui.normalWindowSize.Y)/gtx.Metric.PxPerDp)))
							window.Perform(system.ActionMove)
						}
					}
				}
				event.Op(gtx.Ops, &ui.titleBarTag)
			} else {
				for {
					ev, ok := gtx.Event(pointer.Filter{
						Target: &ui.titleBarTag,
						Kinds:  pointer.Press,
					})
					if !ok {
						break
					}

					if pe, ok := ev.(pointer.Event); ok && pe.Kind == pointer.Press {
						now := gtx.Now.UnixMilli()
						if now-ui.lastTitlePress < 400 {
							window.Perform(system.ActionMaximize)
							ui.lastTitlePress = 0
						} else {
							ui.lastTitlePress = now
						}
					}
				}
				event.Op(gtx.Ops, &ui.titleBarTag)
				ui.decorations.LayoutMove(gtx, func(gtx C) D {
					return D{Size: areaSize}
				})
			}

			area.Pop()

			return D{Size: areaSize}
		}),
		layout.Rigid(func(gtx C) D {
			return ui.layoutWindowButton(gtx, &ui.minimizeBtn, "\u2212", color.NRGBA{R: 160, G: 165, B: 175, A: 255})
		}),
		layout.Rigid(func(gtx C) D {
			symbol := "\u25A1"
			if ui.decorations.Maximized {
				symbol = "\u2750"
			}
			return ui.layoutWindowButton(gtx, &ui.maximizeBtn, symbol, color.NRGBA{R: 160, G: 165, B: 175, A: 255})
		}),
		layout.Rigid(func(gtx C) D {
			return ui.layoutWindowButton(gtx, &ui.closeBtn, "\u00D7", color.NRGBA{R: 160, G: 165, B: 175, A: 255})
		}),
	)
}

func (ui *LauncherUI) layoutWindowButton(gtx C, btn *widget.Clickable, symbol string, textColor color.NRGBA) D {
	width := gtx.Dp(46)
	height := gtx.Dp(40)

	bgColor := color.NRGBA{A: 0}
	if btn.Hovered() {
		if symbol == "\u00D7" {
			bgColor = color.NRGBA{R: 232, G: 17, B: 35, A: 255}
			textColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		} else {
			bgColor = color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		}
	}

	return btn.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min = image.Point{X: width, Y: height}
		gtx.Constraints.Max = gtx.Constraints.Min

		if bgColor.A > 0 {
			bounds := image.Rect(0, 0, width, height)
			paint.FillShape(gtx.Ops, bgColor, clip.Rect(bounds).Op())
		}

		return layout.Center.Layout(gtx, func(gtx C) D {
			lbl := material.Label(ui.theme, unit.Sp(16), symbol)
			lbl.Color = textColor
			return lbl.Layout(gtx)
		})
	})
}

func (ui *LauncherUI) layoutCenterArea(gtx C) D {
	rect := clip.Rect{Max: gtx.Constraints.Max}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 16, G: 16, B: 19, A: 255}, rect.Op())

	return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx C) D {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return ui.layoutPatchlineSelector(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			layout.Rigid(func(gtx C) D {
				return ui.layoutVersionSelector(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			layout.Rigid(func(gtx C) D {
				return ui.layoutInstallLocation(gtx)
			}),
		)
	})
}

func (ui *LauncherUI) layoutPatchlineSelector(gtx C) D {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lbl := material.Label(ui.theme, unit.Sp(14), "Patchline:")
			lbl.Color = color.NRGBA{R: 160, G: 165, B: 175, A: 255}
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					selected := wCommune.Patchline == "release"
					return ui.layoutToggleButton(gtx, &ui.patchlineRelease, "Release", selected)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx C) D {
					selected := wCommune.Patchline == "pre-release"
					return ui.layoutToggleButton(gtx, &ui.patchlinePreRelease, "Pre-Release", selected)
				}),
			)
		}),
	)
}

func (ui *LauncherUI) layoutToggleButton(gtx C, btn *widget.Clickable, text string, selected bool) D {
	bgColor := color.NRGBA{R: 55, G: 60, B: 72, A: 255}
	textColor := color.NRGBA{R: 160, G: 165, B: 175, A: 255}

	if selected {
		bgColor = color.NRGBA{R: 40, G: 113, B: 168, A: 255}
		textColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	} else if btn.Hovered() {
		bgColor = color.NRGBA{R: 65, G: 70, B: 82, A: 255}
	}

	return btn.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
				rrect := clip.RRect{Rect: bounds, SE: 6, SW: 6, NW: 6, NE: 6}
				paint.FillShape(gtx.Ops, bgColor, rrect.Op(gtx.Ops))
				return D{Size: bounds.Size()}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.Inset{
					Left:   unit.Dp(16),
					Right:  unit.Dp(16),
					Top:    unit.Dp(10),
					Bottom: unit.Dp(10),
				}.Layout(gtx, func(gtx C) D {
					lbl := material.Label(ui.theme, unit.Sp(13), text)
					lbl.Color = textColor
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (ui *LauncherUI) layoutVersionSelector(gtx C) D {
	ui.updateVersionButtons()

	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lbl := material.Label(ui.theme, unit.Sp(14), "Version:")
			lbl.Color = color.NRGBA{R: 160, G: 165, B: 175, A: 255}
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return ui.layoutVersionButtons(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
				layout.Rigid(func(gtx C) D {
					return ui.layoutDeleteButton(gtx)
				}),
			)
		}),
	)
}

func (ui *LauncherUI) layoutVersionButtons(gtx C) D {
	latest := wCommune.LatestVersions[wCommune.Patchline]
	children := make([]layout.FlexChild, 0, latest*2)

	for i := 0; i < latest && i < len(ui.versionButtons); i++ {
		version := i + 1
		installed := isGameVersionInstalled(version, wCommune.Patchline)
		selected := wCommune.SelectedVersion == version

		idx := i
		children = append(children, layout.Rigid(func(gtx C) D {
			return ui.layoutVersionButton(gtx, &ui.versionButtons[idx], version, installed, selected)
		}))

		if i < latest-1 {
			children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
		}
	}

	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
}

func (ui *LauncherUI) layoutVersionButton(gtx C, btn *widget.Clickable, version int, installed bool, selected bool) D {
	bgColor := color.NRGBA{R: 55, G: 60, B: 72, A: 255}
	textColor := color.NRGBA{R: 160, G: 165, B: 175, A: 255}
	statusColor := color.NRGBA{R: 255, G: 100, B: 50, A: 255}

	if installed {
		statusColor = color.NRGBA{R: 76, G: 217, B: 100, A: 255}
	}

	if selected {
		bgColor = color.NRGBA{R: 40, G: 113, B: 168, A: 255}
		textColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	} else if btn.Hovered() {
		bgColor = color.NRGBA{R: 65, G: 70, B: 82, A: 255}
	}

	return btn.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
				rrect := clip.RRect{Rect: bounds, SE: 6, SW: 6, NW: 6, NE: 6}
				paint.FillShape(gtx.Ops, bgColor, rrect.Op(gtx.Ops))
				return D{Size: bounds.Size()}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.Inset{
					Left:   unit.Dp(12),
					Right:  unit.Dp(12),
					Top:    unit.Dp(8),
					Bottom: unit.Dp(8),
				}.Layout(gtx, func(gtx C) D {
					return layout.Flex{
						Axis:      layout.Vertical,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							lbl := material.Label(ui.theme, unit.Sp(13), "v"+strconv.Itoa(version))
							lbl.Color = textColor
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(2)}.Layout),
						layout.Rigid(func(gtx C) D {
							statusText := "Not Installed"
							if installed {
								statusText = "Installed"
							}
							lbl := material.Label(ui.theme, unit.Sp(9), statusText)
							lbl.Color = statusColor
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						}),
					)
				})
			}),
		)
	})
}

func (ui *LauncherUI) layoutDeleteButton(gtx C) D {
	installed := isGameVersionInstalled(wCommune.SelectedVersion, wCommune.Patchline)
	if !installed || wDisabled {
		return D{}
	}

	bgColor := color.NRGBA{R: 200, G: 60, B: 60, A: 255}
	if ui.deleteBtn.Hovered() {
		bgColor = color.NRGBA{R: 220, G: 80, B: 80, A: 255}
	}

	return ui.deleteBtn.Layout(gtx, func(gtx C) D {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
				rrect := clip.RRect{Rect: bounds, SE: 6, SW: 6, NW: 6, NE: 6}
				paint.FillShape(gtx.Ops, bgColor, rrect.Op(gtx.Ops))
				return D{Size: bounds.Size()}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.Inset{
					Left:   unit.Dp(16),
					Right:  unit.Dp(16),
					Top:    unit.Dp(14),
					Bottom: unit.Dp(14),
				}.Layout(gtx, func(gtx C) D {
					lbl := material.Label(ui.theme, unit.Sp(12), "Delete")
					lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (ui *LauncherUI) layoutInstallLocation(gtx C) D {
	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			lbl := material.Label(ui.theme, unit.Sp(14), "Install Location:")
			lbl.Color = color.NRGBA{R: 160, G: 165, B: 175, A: 255}
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx C) D {
					return ui.layoutPathInput(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx C) D {
					return ui.layoutBrowseButton(gtx)
				}),
			)
		}),
	)
}

func (ui *LauncherUI) layoutPathInput(gtx C) D {
	height := gtx.Dp(40)

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			gtx.Constraints.Min.Y = height
			gtx.Constraints.Max.Y = height
			bounds := image.Rect(0, 0, gtx.Constraints.Max.X, height)
			rrect := clip.RRect{Rect: bounds, SE: 6, SW: 6, NW: 6, NE: 6}
			paint.FillShape(gtx.Ops, color.NRGBA{R: 50, G: 54, B: 64, A: 255}, rrect.Op(gtx.Ops))
			return D{Size: image.Point{X: gtx.Constraints.Max.X, Y: height}}
		}),
		layout.Stacked(func(gtx C) D {
			return layout.Inset{
				Left:  unit.Dp(12),
				Right: unit.Dp(12),
				Top:   unit.Dp(10),
			}.Layout(gtx, func(gtx C) D {
				editor := material.Editor(ui.theme, &ui.installPathField, "Install path...")
				editor.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				editor.HintColor = color.NRGBA{R: 160, G: 165, B: 175, A: 255}
				editor.TextSize = unit.Sp(14)
				return editor.Layout(gtx)
			})
		}),
	)
}

func (ui *LauncherUI) layoutBrowseButton(gtx C) D {
	bgColor := color.NRGBA{R: 55, G: 60, B: 72, A: 255}
	if ui.browseBtn.Hovered() {
		bgColor = color.NRGBA{R: 65, G: 70, B: 82, A: 255}
	}

	return ui.browseBtn.Layout(gtx, func(gtx C) D {
		return layout.Stack{Alignment: layout.Center}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				bounds := image.Rect(0, 0, gtx.Constraints.Min.X, gtx.Constraints.Min.Y)
				rrect := clip.RRect{Rect: bounds, SE: 6, SW: 6, NW: 6, NE: 6}
				paint.FillShape(gtx.Ops, bgColor, rrect.Op(gtx.Ops))
				return D{Size: bounds.Size()}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.Inset{
					Left:   unit.Dp(16),
					Right:  unit.Dp(16),
					Top:    unit.Dp(12),
					Bottom: unit.Dp(12),
				}.Layout(gtx, func(gtx C) D {
					lbl := material.Label(ui.theme, unit.Sp(12), "Browse")
					lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (ui *LauncherUI) layoutBottomBar(gtx C) D {
	height := gtx.Dp(80)
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	rect := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: height}}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 00, G: 00, B: 00, A: 255}, rect.Op())

	return layout.Inset{
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
		Top:    unit.Dp(16),
		Bottom: unit.Dp(16),
	}.Layout(gtx, func(gtx C) D {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceBetween,
		}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return ui.layoutUsernameInput(gtx)
			}),
			layout.Flexed(1, func(gtx C) D {
				return D{}
			}),
			layout.Rigid(func(gtx C) D {
				return ui.layoutPlayButtonCombo(gtx)
			}),
		)
	})
}

func (ui *LauncherUI) layoutUsernameInput(gtx C) D {
	width := gtx.Dp(250)
	height := gtx.Dp(48)

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			bounds := image.Rect(0, 0, width, height)
			rrect := clip.RRect{Rect: bounds, SE: 8, SW: 8, NW: 8, NE: 8}
			paint.FillShape(gtx.Ops, color.NRGBA{R: 50, G: 54, B: 64, A: 255}, rrect.Op(gtx.Ops))
			return D{Size: image.Point{X: width, Y: height}}
		}),
		layout.Stacked(func(gtx C) D {
			gtx.Constraints.Min = image.Point{X: width, Y: height}
			gtx.Constraints.Max = gtx.Constraints.Min

			return layout.Inset{
				Left:  unit.Dp(16),
				Right: unit.Dp(16),
			}.Layout(gtx, func(gtx C) D {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx C) D {
						editor := material.Editor(ui.theme, &ui.usernameField, "Username")
						editor.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
						editor.HintColor = color.NRGBA{R: 160, G: 165, B: 175, A: 255}
						editor.TextSize = unit.Sp(16)
						return layout.Center.Layout(gtx, editor.Layout)
					}),
				)
			})
		}),
	)
}

func (ui *LauncherUI) layoutPlayButtonCombo(gtx C) D {
	height := gtx.Dp(48)
	playWidth := gtx.Dp(120)
	dropdownWidth := gtx.Dp(40)

	blueColor := color.NRGBA{R: 40, G: 113, B: 168, A: 255}
	blueColorHover := color.NRGBA{R: 1, G: 77, B: 102, A: 255}

	if ui.playBtn.Hovered() {
		blueColor = blueColorHover
	}
	if ui.playDropdown.Hovered() {
		blueColor = blueColorHover
	}
	if wDisabled {
		blueColor = color.NRGBA{R: 55, G: 60, B: 72, A: 255}
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return ui.playBtn.Layout(gtx, func(gtx C) D {
						bounds := image.Rect(0, 0, playWidth, height)
						rrect := clip.RRect{Rect: bounds, NW: 8, SW: 8}
						paint.FillShape(gtx.Ops, blueColor, rrect.Op(gtx.Ops))

						gtx.Constraints.Min = image.Point{X: playWidth, Y: height}
						gtx.Constraints.Max = gtx.Constraints.Min

						return layout.Center.Layout(gtx, func(gtx C) D {
							text := "PLAY"
							if wDisabled {
								text = fmt.Sprintf("%d%%", wProgress)
							}
							lbl := material.Label(ui.theme, unit.Sp(16), text)
							lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						})
					})
				}),
				layout.Rigid(func(gtx C) D {
					sepWidth := gtx.Dp(1.5)
					bounds := image.Rect(0, 0, sepWidth, height)
					paint.FillShape(gtx.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 0}, clip.Rect(bounds).Op())
					return D{Size: image.Point{X: sepWidth, Y: height}}
				}),
				layout.Rigid(func(gtx C) D {
					return ui.playDropdown.Layout(gtx, func(gtx C) D {
						bounds := image.Rect(0, 0, dropdownWidth, height)
						rrect := clip.RRect{Rect: bounds, NE: 8, SE: 8}
						paint.FillShape(gtx.Ops, blueColor, rrect.Op(gtx.Ops))

						gtx.Constraints.Min = image.Point{X: dropdownWidth, Y: height}
						gtx.Constraints.Max = gtx.Constraints.Min

						return layout.Center.Layout(gtx, func(gtx C) D {
							symbol := "\u25BC"
							if ui.dropdownOpen {
								symbol = "\u25B2"
							}
							lbl := material.Label(ui.theme, unit.Sp(12), symbol)
							lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
							return lbl.Layout(gtx)
						})
					})
				}),
			)
		}),
		layout.Expanded(func(gtx C) D {
			if ui.dropdownOpen {
				return ui.layoutModeDropdown(gtx)
			}
			return D{}
		}),
	)
}

func (ui *LauncherUI) layoutModeDropdown(gtx C) D {
	dropdownHeight := gtx.Dp(80)
	totalWidth := gtx.Dp(161)

	return layout.Inset{Top: unit.Dp(-80)}.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx C) D {
				bounds := image.Rect(0, 0, totalWidth, dropdownHeight)
				rrect := clip.RRect{Rect: bounds, NW: 8, NE: 8, SW: 0, SE: 0}
				paint.FillShape(gtx.Ops, color.NRGBA{R: 55, G: 60, B: 72, A: 255}, rrect.Op(gtx.Ops))
				return D{Size: image.Point{X: totalWidth, Y: dropdownHeight}}
			}),
			layout.Stacked(func(gtx C) D {
				gtx.Constraints.Min = image.Point{X: totalWidth, Y: dropdownHeight}
				gtx.Constraints.Max = gtx.Constraints.Min

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return ui.layoutModeOption(gtx, &ui.modeOffline, "Offline", wCommune.Mode == "offline")
					}),
					layout.Rigid(func(gtx C) D {
						return ui.layoutModeOption(gtx, &ui.modeFakeOnline, "Fake Online", wCommune.Mode == "fakeonline")
					}),
				)
			}),
		)
	})
}

func (ui *LauncherUI) layoutModeOption(gtx C, btn *widget.Clickable, text string, selected bool) D {
	height := gtx.Dp(40)

	bgColor := color.NRGBA{A: 0}
	textColor := color.NRGBA{R: 160, G: 165, B: 175, A: 255}
	if selected {
		textColor = color.NRGBA{R: 40, G: 113, B: 168, A: 255}
	}
	if btn.Hovered() {
		bgColor = color.NRGBA{R: 70, G: 75, B: 85, A: 255}
	}

	return btn.Layout(gtx, func(gtx C) D {
		gtx.Constraints.Min.Y = height
		gtx.Constraints.Max.Y = height

		if bgColor.A > 0 {
			bounds := image.Rect(0, 0, gtx.Constraints.Max.X, height)
			paint.FillShape(gtx.Ops, bgColor, clip.Rect(bounds).Op())
		}

		return layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					checkmark := ""
					if selected {
						checkmark = "\u2713 "
					}
					lbl := material.Label(ui.theme, unit.Sp(12), checkmark+text)
					lbl.Color = textColor
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (ui *LauncherUI) handleEvents(gtx C, window *app.Window) {
	if ui.patchlineRelease.Clicked(gtx) && wCommune.Patchline != "release" {
		wCommune.Patchline = "release"
		wCommune.SelectedVersion = wCommune.LatestVersions["release"]
		ui.updateVersionButtons()
	}
	if ui.patchlinePreRelease.Clicked(gtx) && wCommune.Patchline != "pre-release" {
		wCommune.Patchline = "pre-release"
		wCommune.SelectedVersion = wCommune.LatestVersions["pre-release"]
		ui.updateVersionButtons()
	}

	for i := range ui.versionButtons {
		if ui.versionButtons[i].Clicked(gtx) {
			wCommune.SelectedVersion = i + 1
		}
	}

	if ui.modeOffline.Clicked(gtx) {
		wCommune.Mode = "offline"
		ui.dropdownOpen = false
	}
	if ui.modeFakeOnline.Clicked(gtx) {
		wCommune.Mode = "fakeonline"
		ui.dropdownOpen = false
	}

	if ui.playDropdown.Clicked(gtx) {
		ui.dropdownOpen = !ui.dropdownOpen
	}

	if ui.browseBtn.Clicked(gtx) {
		go func() {
			dir, err := dialog.Directory().Title("Select install location").Browse()
			if err == nil && dir != "" {
				wCommune.GameFolder = dir
				ui.installPathField.SetText(dir)
				window.Invalidate()
			}
		}()
	}

	if ui.deleteBtn.Clicked(gtx) && !wDisabled {
		go func() {
			wDisabled = true
			window.Invalidate()

			installDir := getVersionInstallPath(wCommune.SelectedVersion, wCommune.Patchline)
			err := os.RemoveAll(installDir)
			if err != nil {
				fmt.Printf("Failed to remove: %s\n", err)
			}

			wDisabled = false
			window.Invalidate()
		}()
	}

	if ui.playBtn.Clicked(gtx) && !wDisabled {
		go func() {
			startGame()
			window.Invalidate()
		}()
	}

	if ui.closeBtn.Clicked(gtx) {
		writeSettings()
		os.Exit(0)
	}
	if ui.minimizeBtn.Clicked(gtx) {
		window.Perform(system.ActionMinimize)
	}
	if ui.maximizeBtn.Clicked(gtx) {
		if ui.decorations.Maximized {
			window.Perform(system.ActionUnmaximize)
		} else {
			window.Perform(system.ActionMaximize)
		}
	}
}

func runGioUI() error {
	ui := NewLauncherUI()

	go func() {
		window := new(app.Window)
		window.Option(
			app.Title("hytLauncher"),
			app.Size(unit.Dp(900), unit.Dp(600)),
			app.MinSize(unit.Dp(800), unit.Dp(500)),
			app.Decorated(false),
		)

		var ops op.Ops
		for {
			switch e := window.Event().(type) {
			case app.DestroyEvent:
				writeSettings()
				return
			case app.ConfigEvent:
				if e.Config.Mode != app.Maximized && e.Config.Size.X > 0 {
					ui.normalWindowSize = e.Config.Size
				}
				ui.decorations.Maximized = e.Config.Mode == app.Maximized
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				if actions := ui.decorations.Update(gtx); actions != 0 {
					window.Perform(actions)
				}

				ui.handleEvents(gtx, window)

				ui.Layout(gtx, window)
				e.Frame(gtx.Ops)
			}
		}
	}()

	app.Main()
	return nil
}

func startGame() {
	wDisabled = true
	installJre(updateProgress)
	installGame(wCommune.SelectedVersion, wCommune.Patchline, updateProgress)
	launchGame(wCommune.SelectedVersion, wCommune.Patchline, wCommune.Username, usernameToUuid(wCommune.Username))
	wDisabled = false
}
