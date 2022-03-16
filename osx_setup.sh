#!/usr/bin/env bash
echo "Starting batch installation"

# Homebrew Script for OSX
# To execute: save and `chmod +x ./brew-install-script.sh` then `./brew-install-script.sh`

if [[ -z "${CI}" ]]; then
  sudo -v # Ask for the administrator password upfront
  # Keep-alive: update existing `sudo` time stamp until script has finished
  while true; do sudo -n true; sleep 60; kill -0 "$$" || exit; done 2>/dev/null &
fi

# Create develoepr folder
mkdir -p ~/Developer

echo "Checking brew..."
if test ! $(which brew); then
    echo "Installing homebrew..."
    ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
fi

echo "Installing brew cask..."
brew tap homebrew/cask

# Update homebrew recipes
brew update

# Programming Languages
echo "Installing programming languages..."
brew install --cask r
brew install --cask python
brew install --cask python3
brew install --cask oracle-jdk-javadoc
brew install --cask adoptopenjdk8
brew install --cask adoptopenjdk11
brew install --cask adoptopenjdk13
brew install --cask java
brew tap adoptopenjdk/openjdk
brew install go
brew install postgresql

# Dev Tools
echo "Installing development tools..."
brew install --cask docker
brew install --cask github
brew install --cask hyper
brew install --cask iterm2
brew install --cask kitematic
brew install --cask rstudio
brew install --cask visual-studio-code
brew install --cask atom
brew install --cask postman
brew install docker
brew install git
brew install vim
brew install tmux
brew install tree
brew install wget

echo "Installing VSCode extensions..."
code --install-extension msjsdiag.debugger-for-chrome
code --install-extension hookyqr.beautify
code --install-extension dbaeumer.vscode-eslint
code --install-extension eamodio.gitlens
code --install-extension ritwickdey.liveserver
code --install-extension fallenmax.mithril-emmet
code --install-extension wayou.vscode-todo-highlight


echo "Git config"
git config --global user.name "AngeloAndreaIsola"
git config --global user.email andrea.isola@me.com


# Communication Apps
echo "Installing communication apps..."
brew install --cask discord
brew install --cask microsoft-teams
brew install --cask skype
brew install --cask whatsapp
brew install --cask telegram
brew install --cask zoom
brew install --cask spark

# Web Tools
echo "Installing web tools..."
brew install node
brew install nvm
brew install --cask firefox
brew install --cask google-chrome

# File Storage
echo "Installing file storage tools..."
brew install --cask dropbox
brew install --cask onedrive
brew install --cask google-drive

# Writing Apps
echo "Installing writing apps..."
brew install --cask microsoft-word
brew install --cask microsoft-excel
brew install --cask microsoft-powerpoint

# Other
echo "Installing everything else..."
brew install --cask anki
brew install --cask lastpass
brew install --cask the-unarchiver
brew install --cask teamviewer
brew install --cask vlc
brew install --cask transmission
brew install --cask adobe-acrobat-reader
brew install --cask 4k-video-downloader
brew install --cask copyclip
brew install --cask alfred
brew install --cask android-file-transfer
brew install --cask monitorcontrol
brew install --cask rectangle
brew install --cask soundflower
brew install --cask bartender

echo "Cleaning up..."
brew cleanup

echo "Enabling Services"
open /Applications/Alfred\ 4.app

echo "Configuring OSX..."
# Show filename extensions by default
defaults write NSGlobalDomain AppleShowAllExtensions -bool true

#"Allow text selection in Quick Look"
defaults write com.apple.finder QLEnableTextSelection -bool TRUE

# Show the ~/Library folder
chflags nohidden ~/Library

# Show icons for hard drives, servers, and removable media on the desktop
defaults write com.apple.finder ShowExternalHardDrivesOnDesktop -bool true
defaults write com.apple.finder ShowHardDrivesOnDesktop -bool true
defaults write com.apple.finder ShowMountedServersOnDesktop -bool true
defaults write com.apple.finder ShowRemovableMediaOnDesktop -bool true

#"Avoiding the creation of .DS_Store files on network volumes"
defaults write com.apple.desktopservices DSDontWriteNetworkStores -bool true

#"Setting Dock to auto-hide and removing the auto-hiding delay"
defaults write com.apple.dock autohide -bool true
defaults write com.apple.dock autohide-delay -float 0
defaults write com.apple.dock autohide-time-modifier -float 0

#"Setting email addresses to copy as 'foo@example.com' instead of 'Foo Bar <foo@example.com>' in Mail.app"
defaults write com.apple.mail AddressesIncludeNameOnPasteboard -bool false

#"Setting screenshots location to ~/Desktop"
defaults write com.apple.screencapture location -string "$HOME/Desktop"

#"Setting screenshot format to PNG"
defaults write com.apple.screencapture type -string "png"

# Finder: show status bar
defaults write com.apple.finder ShowStatusBar -bool true

# Finder: show path bar
defaults write com.apple.finder ShowPathbar -bool true

# Save to disk (not to iCloud) by default
defaults write NSGlobalDomain NSDocumentSaveNewDocumentsToCloud -bool false

# Enable full keyboard access for all controls
# (e.g. enable Tab in modal dialogs)
defaults write NSGlobalDomain AppleKeyboardUIMode -int 3
