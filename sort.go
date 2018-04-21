// Copyright © 2018, Lukas Werling
// Copyright (c) 1998-2000, Matthes Bender
// Copyright (c) 2001-2009, RedWolf Design GmbH, http://www.clonk.de/
// Copyright (c) 2009-2016, The OpenClonk Team and contributors
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package c4group

import (
	"path/filepath"
	"strings"
)

// Groups are sorted to speed up loading. Sort data is from c4group/C4Components.h

// Fix #define:
// :%s/\v^#define (\w+)\s*/\t\1 =
// Fix some constant interpolation:
// :%s/\v\|"\s*(C4\w+)/|"+\1
// :%s/\v(C4\w+)\s*"/\1+"
// Fix line continuations.
// :%s/\\$/+

// Component file names
const (
	C4CFN_Material       = "Material.ocg"
	C4CFN_Sound          = "Sound.ocg"
	C4CFN_SoundSubgroups = "*.ocg"
	C4CFN_Graphics       = "Graphics.ocg"
	C4CFN_System         = "System.ocg"
	C4CFN_Music          = "Music.ocg"
	C4CFN_Extra          = "Extra.ocg"
	C4CFN_Languages      = "Language.ocg"
	C4CFN_Template       = "Template.ocg"

	C4CFN_Savegames = "Savegames.ocf"
	C4CFN_Records   = "Records.ocf"

	C4CFN_ScenarioSections = "Sect*.ocg"

	C4CFN_Objects = "Objects.ocd"

	C4CFN_ScenarioCore          = "Scenario.txt"
	C4CFN_ScenarioParameterDefs = "ParameterDefs.txt"
	C4CFN_FolderCore            = "Folder.txt"
	C4CFN_PlayerInfoCore        = "Player.txt"
	C4CFN_DefCore               = "DefCore.txt"
	C4CFN_ObjectInfoCore        = "ObjectInfo.txt"
	C4CFN_ParticleCore          = "Particle.txt"
	C4CFN_LinkCore              = "Link.txt"
	C4CFN_UpdateCore            = "AutoUpdate.txt"
	C4CFN_UpdateEntries         = "GRPUP_Entries.txt"

	C4CFN_UpdateGroupExtension = ".ocu"
	C4CFN_UpdateProgram        = "c4group"
	C4CFN_UpdateProgramLibs    = "*.dll"

	C4CFN_Map                   = "Map.bmp"
	C4CFN_MapFg                 = "MapFg.bmp"
	C4CFN_MapBg                 = "MapBg.bmp"
	C4CFN_Landscape             = "Landscape.bmp"
	C4CFN_LandscapeFg           = "LandscapeFg.bmp"
	C4CFN_LandscapeBg           = "LandscapeBg.bmp"
	C4CFN_DiffLandscape         = "DiffLandscape.bmp"
	C4CFN_DiffLandscapeBkg      = "DiffLandscapeBkg.bmp"
	C4CFN_Sky                   = "Sky"
	C4CFN_Script                = "Script.c|Script%s.c|C4Script%s.c"
	C4CFN_MapScript             = "Map.c"
	C4CFN_ScriptStringTbl       = "StringTbl.txt|StringTbl%s.txt"
	C4CFN_AnyScriptStringTbl    = "StringTbl*.txt"
	C4CFN_Info                  = "Info.txt"
	C4CFN_Author                = "Author.txt"
	C4CFN_Version               = "Version.txt"
	C4CFN_Game                  = "Game.txt"
	C4CFN_ScenarioObjectsScript = "Objects.c"
	C4CFN_PXS                   = "PXS.ocb"
	C4CFN_MassMover             = "MassMover.ocb"
	C4CFN_CtrlRec               = "CtrlRec.ocb"
	C4CFN_CtrlRecText           = "CtrlRec.txt"
	C4CFN_LogRec                = "Record.log"
	C4CFN_TexMap                = "TexMap.txt"
	C4CFN_MatMap                = "MatMap.txt"
	C4CFN_Title                 = "Title%s.txt|Title.txt"
	C4CFN_WriteTitle            = "Title.txt" // file that is generated if a title is set automatically
	C4CFN_ScenarioTitle         = "Title"
	C4CFN_ScenarioIcon          = "Icon.bmp"
	C4CFN_IconPNG               = "Icon.png"
	C4CFN_ScenarioObjects       = "Objects.txt"
	C4CFN_ScenarioDesc          = "Desc%s.txt"
	C4CFN_DefMaterials          = "*.material"
	C4CFN_Achievements          = "Achv*.png"

	C4CFN_DefMesh              = "Graphics.mesh"
	C4CFN_DefMeshXml           = C4CFN_DefMesh + ".xml"
	C4CFN_DefSkeleton          = "*.skeleton"
	C4CFN_DefSkeletonXml       = C4CFN_DefSkeleton + ".xml"
	C4CFN_DefGraphicsExMesh    = "Graphics*.mesh"
	C4CFN_DefGraphicsExMeshXml = C4CFN_DefGraphicsExMesh + ".xml"

	C4CFN_DefGraphics   = "Graphics.png"
	C4CFN_ClrByOwner    = "Overlay.png"
	C4CFN_NormalMap     = "Normal.png"
	C4CFN_DefGraphicsEx = "Graphics*.png"
	C4CFN_ClrByOwnerEx  = "Overlay*.png"
	C4CFN_NormalMapEx   = "Normal*.png"

	C4CFN_DefGraphicsScaled = "Graphics.*.png"
	C4CFN_ClrByOwnerScaled  = "Graphics.*.png"
	C4CFN_NormalMapScaled   = "Normal.*.png"

	C4CFN_DefDesc            = "Desc%s.txt"
	C4CFN_BigIcon            = "BigIcon.png"
	C4CFN_UpperBoard         = "UpperBoard"
	C4CFN_Logo               = "Logo"
	C4CFN_MoreMusic          = "MoreMusic.txt"
	C4CFN_DynLandscape       = "Landscape.txt"
	C4CFN_ClonkNames         = "ClonkNames%s.txt|ClonkNames.txt"
	C4CFN_ClonkNameFiles     = "ClonkNames*.txt"
	C4CFN_RankNames          = "Rank%s.txt|Rank.txt"
	C4CFN_RankNameFiles      = "Rank*.txt"
	C4CFN_RankFacesPNG       = "Rank.png"
	C4CFN_ClonkRank          = "Rank.png" // custom rank in info file: One rank image only
	C4CFN_SolidMask          = "SolidMask.png"
	C4CFN_LeagueInfo         = "League.txt" // read by frontend only
	C4CFN_PlayerInfos        = "PlayerInfos.txt"
	C4CFN_SavePlayerInfos    = "SavePlayerInfos.txt"
	C4CFN_RecPlayerInfos     = "RecPlayerInfos.txt"
	C4CFN_Teams              = "Teams.txt"
	C4CFN_Parameters         = "Parameters.txt"
	C4CFN_RoundResults       = "RoundResults.txt"
	C4CFN_PlayerControls     = "PlayerControls.txt"
	C4CFN_LandscapeShader    = "LandscapeShader.c"
	C4CFN_LandscapeScaler    = "Scaler.png"
	C4CFN_MaterialShapeFiles = "_Shape.png"

	C4CFN_MapFolderData = "FolderMap.txt"
	C4CFN_MapFolderBG   = "FolderMap"

	C4CFN_Language  = "Language*.txt"
	C4CFN_KeyConfig = "KeyConfig.txt"

	C4CFN_Log                     = "OpenClonk.log"
	C4CFN_LogEx                   = "OpenClonk%d.log"      // created if regular logfile is in use
	C4CFN_LogShader               = "OpenClonkShaders.log" // created in editor mode to dump shader code
	C4CFN_Intro                   = "Clonk4.avi"
	C4CFN_Names                   = "Names.txt"
	C4CFN_Titles                  = "Title*.txt|Title.txt"
	C4CFN_DefNameFiles            = "Names*.txt|Names.txt"
	C4CFN_EditorGeometry          = "Editor.geometry"
	C4CFN_DefaultScenarioTemplate = "Empty.ocs"

	C4CFN_TempMusic        = "~Music.tmp"
	C4CFN_TempMusic2       = "~Music2.tmp"
	C4CFN_TempSky          = "~Sky.tmp"
	C4CFN_TempMapFg        = "~MapFg.tmp"
	C4CFN_TempMapBg        = "~MapBg.tmp"
	C4CFN_TempLandscape    = "~Landscape.tmp"
	C4CFN_TempLandscapeBkg = "~LandscapeBkg.tmp"
	C4CFN_TempPXS          = "~PXS.tmp"
	C4CFN_TempTitle        = "~Title.tmp"
	C4CFN_TempCtrlRec      = "~CtrlRec.tmp"
	C4CFN_TempReSync       = "~ReSync.tmp"
	C4CFN_TempPlayer       = "~plr.tmp"
	C4CFN_TempRoundResults = "~C4Results.tmp"
	C4CFN_TempLeagueInfo   = "~league.tmp"

	C4CFN_DefFiles          = "*.ocd"
	C4CFN_GenericGroupFiles = "*.ocg"
	C4CFN_PlayerFiles       = "*.ocp"
	C4CFN_MaterialFiles     = "*.ocm"
	C4CFN_ObjectInfoFiles   = "*.oci"
	C4CFN_MusicFiles        = "*.ogg"
	C4CFN_SoundFiles        = "*.wav|*.ogg"
	C4CFN_PNGFiles          = "*.png"
	C4CFN_BitmapFiles       = "*.bmp"
	C4CFN_ScenarioFiles     = "*.ocs"
	C4CFN_FolderFiles       = "*.ocf"
	C4CFN_QueueFiles        = "*.c4q"
	C4CFN_AnimationFiles    = "*.ocv"
	C4CFN_KeyFiles          = "*.c4k"
	C4CFN_ScriptFiles       = "*.c"
	C4CFN_ImageFiles        = "*.png|*.bmp|*.jpeg|*.jpg"
	C4CFN_FontFiles         = "*.fon|*.fnt|*.ttf|*.ttc|*.fot|*.otf"
	C4CFN_ShaderFiles       = "*.glsl"
)

// Sort lists
const (
	C4FLS_Scenario         = "Loader*.bmp|Loader*.png|Loader*.jpeg|Loader*.jpg|Fonts.txt|Scenario.txt|Title*.txt|Info.txt|Desc*.txt|Icon.png|Icon.bmp|Achv*.png|Game.txt|StringTbl*.txt|ParameterDefs.txt|Teams.txt|Parameters.txt|Info.txt|Sect*.ocg|Music.ocg|*.mid|*.wav|Desc*.txt|Title.png|Title.jpg|*.ocd|Script.c|Script*.c|Map.c|Objects.c|System.ocg|Material.ocg|MatMap.txt|Map.bmp|MapFg.bmp|MapBg.bmp|Landscape.bmp|LandscapeFg.bmp|LandscapeBg.bmp|" + C4CFN_DiffLandscape + "|" + C4CFN_DiffLandscapeBkg + "|Sky.bmp|Sky.png|Sky.jpeg|Sky.jpg|PXS.ocb|MassMover.ocb|CtrlRec.ocb|Strings.txt|Objects.txt|RoundResults.txt|Author.txt|Version.txt|Names.txt"
	C4FLS_Section          = "Scenario.txt|Game.txt|Map.bmp|MapFg.bmp|MapBg.bmp|Landscape.bmp|LandscapeFg.bmp|LandscapeBg.bmp|Sky.bmp|Sky.png|Sky.jpeg|Sky.jpg|PXS.ocb|MassMover.ocb|CtrlRec.ocb|Strings.txt|Objects.txt|Objects.c"
	C4FLS_SectionLandscape = "Scenario.txt|Map.bmp|MapFg.bmp|MapBg.bmp|Landscape.bmp|LandscapeFg.bmp|LandscapeBg.bmp|PXS.ocb|MassMover.ocb"
	C4FLS_SectionObjects   = "Strings.txt|Objects.txt|Objects.c"
	C4FLS_Def              = "*.glsl|*.png|*.bmp|*.jpeg|*.jpg|*.material|Particle.txt|DefCore.txt|*.wav|*.ogg|*.skeleton|Graphics.mesh|*.mesh|StringTbl*.txt|Script.c|Script*.c|C4Script.c|Names*.txt|Title*.txt|ClonkNames.txt|Rank.txt|Rank*.txt|Desc*.txt|Author.txt|Version.txt|*.ocd"
	C4FLS_Player           = "Player.txt|BigIcon.png|*.oci"
	C4FLS_Object           = "ObjectInfo.txt"
	C4FLS_Folder           = "Folder.txt|Title*.txt|Info.txt|Desc*.txt|Title.png|Title.jpg|Icon.png|Icon.bmp|Author.txt|Version.txt|StringTbl*.txt|ParameterDefs.txt|Achv*.png|*.ocs|Loader*.bmp|Loader*.png|Loader*.jpeg|Loader*.jpg|FolderMap.txt|FolderMap.png"
	C4FLS_Material         = "TexMap.txt|*.ocm|*.jpeg|*.jpg|*.bmp|*.png"
	C4FLS_Graphics         = "Loader*.bmp|Loader*.png|Loader*.jpeg|Loader*.jpg|*.glsl|Font*.png" +
		"|GUIProgress.png|Endeavour.ttf|GUICaption.png|GUIButton.png|GUIButtonDown.png|GUIButtonHighlight.png|GUIButtonHighlightRound.png|GUIIcons.png|GUIIcons2.png|GUIScroll.png|GUIContext.png|GUISubmenu.png|GUICheckBox.png|GUIBigArrows.png" +
		"|Control.png|ClonkSkins.png|Fire.png|Background.png|Flag.png|Crew.png|Wealth.png|Player.png|Rank.png|Captain.png|Cursor.png|SelectMark.png|MenuSymbol.png|Menu.png|Logo.png|Construction.png|Energy.png|Options.png|UpperBoard.png|Arrow.png|Exit.png|Hand.png|Gamepad.png|Build.png|TransformKnob.png|Achv*.png" +
		"|StartupMainMenuBG.*|StartupScenSelBG.*|StartupPlrSelBG.*|StartupPlrPropBG.*|StartupNetworkBG.*|StartupAboutBG.*|StartupBigButton.png|StartupBigButtonDown.png|StartupBookScroll.png|StartupContext.png|StartupScenSelIcons.png|StartupScenSelTitleOv.png|StartupDlgPaper.png|StartupOptionIcons.png|StartupTabClip.png|StartupNetGetRef.png|StartupLogo.png"
	C4FLS_Objects = "Names*.txt|Desc*.txt|*.ocd"
	C4FLS_System  = "*.hlp|*.cnt|Language*.txt|*.fon|*.fnt|*.ttf|*.ttc|*.fot|*.otf|Fonts.txt|StringTbl*.txt|PlayerControls.txt|*.c|Names.txt"
	C4FLS_Sound   = C4CFN_SoundFiles + "|" + C4CFN_SoundSubgroups
	C4FLS_Music   = C4CFN_MusicFiles
)

var sortLists = []string{
	C4CFN_System, C4FLS_System,
	C4CFN_Material, C4FLS_Material,
	C4CFN_Graphics, C4FLS_Graphics,
	C4CFN_DefFiles, C4FLS_Def,
	C4CFN_PlayerFiles, C4FLS_Player,
	C4CFN_ObjectInfoFiles, C4FLS_Object,
	C4CFN_ScenarioFiles, C4FLS_Scenario,
	C4CFN_FolderFiles, C4FLS_Folder,
	C4CFN_ScenarioSections, C4FLS_Section,
	C4CFN_Sound, C4FLS_Sound,
	C4CFN_Music, C4FLS_Music,
}

// NameLess returns a name sorting function (Less) for a group with the given
// name.
func NameLess(groupName string) func(child1, child2 string) bool {
	var sortList []string
	for i := 0; i < len(sortLists); i += 2 {
		if m, _ := filepath.Match(sortLists[i], groupName); m {
			sortList = strings.Split(sortLists[i+1], "|")
			for j := range sortList {
				sortList[j] = strings.ToLower(sortList[j])
			}
			break
		}
	}
	return func(child1, child2 string) bool {
		child1 = strings.ToLower(child1)
		child2 = strings.ToLower(child2)
		if sortList != nil {
			for _, pattern := range sortList {
				m1, _ := filepath.Match(pattern, child1)
				m2, _ := filepath.Match(pattern, child2)
				if m1 || m2 {
					if m1 && !m2 {
						return true
					}
					if !m1 && m2 {
						return false
					}
					// both in the same group
					break
				}
			}
		}
		return child1 < child2
	}
}
